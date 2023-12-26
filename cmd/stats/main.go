package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/inkel/cabb"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func dieIf(err error) {
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		panic(err)
	}
}

func mustEnv(k string) string {
	v, ok := os.LookupEnv(k)
	if !ok && v == "" {
		panic("missing environment variable " + k)
	}
	return v
}

var (
	deviceID = mustEnv("DEVICEID")
	teamID   = mustEnv("TEAMID")
	defe     = mustEnv("TEAM")
	uid      = mustEnv("CABBUID")
)

const schema = `
CREATE TABLE IF NOT EXISTS teams (
       id TEXT PRIMARY KEY,
       club TEXT NOT NULL,
       name TEXT NOT NULL,
       notificationId TEXT
);

CREATE TABLE IF NOT EXISTS seasons (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       teamId TEXT REFERENCES teams (id) NOT NULL
);

CREATE TABLE IF NOT EXISTS sqlite_sequence(name,seq);

CREATE UNIQUE INDEX IF NOT EXISTS season_team_id ON seasons (teamId);

CREATE TABLE IF NOT EXISTS gamedays (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       seasonId INT REFERENCES seasons (id) NOT NULL,
       name TEXT NOT NULL,
       date TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS match_results (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       matchId TEXT UNIQUE NOT NULL,
       homeTeam TEXT NOT NULL,
       awayTeam TEXT NOT NULL,
       homeScore INTEGER NOT NULL,
       awayScore INTEGER NOT NULL,
       date TEXT NOT NULL,
       status TEXT NOT NULL
);
`

type playerStats struct {
	cabb.PlayerStats
	GamesPlayed int
}

type teamStats map[string]playerStats

type match struct {
	cabb.Match
}

type matches []match

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	dieIf(err)
	return n
}

func (ms matches) Stats(team string) TeamStats {
	var res TeamStats
	for _, m := range ms {
		home, away := atoi(m.HomeScore), atoi(m.AwayScore)
		if m.HomeTeam == team {
			if home > away {
				res.Won += 1
			} else {
				res.Lost += 1
			}
			res.Scored += home
			res.Received += away
		} else {
			if home < away {
				res.Won += 1
			} else {
				res.Lost += 1
			}
			res.Received += home
			res.Scored += away
		}
	}

	return res
}

type TeamStats struct {
	Won, Lost        int
	Scored, Received int
}

type templateData struct {
	Matches     matches
	PlayerStats teamStats

	Team, TeamID string
}

func main() {
	var html bool

	flag.BoolVar(&html, "html", false, "Output")
	flag.Parse()

	c, err := cabb.NewClient(uid, deviceID)
	dieIf(err)

	s, err := c.Season(teamID)
	dieIf(err)

	db, err := sqlx.Connect("sqlite3", "cabb.db")
	dieIf(err)
	defer db.Close()

	var seasonID int64

	err = db.QueryRowx("SELECT id FROM seasons WHERE teamId = $1", teamID).Scan(&seasonID)
	dieIf(err)

	if seasonID == 0 {
		res, err := db.Exec("INSERT INTO seasons (teamId) VALUES ($1)", teamID)
		dieIf(err)

		seasonID, err = res.LastInsertId()
		dieIf(err)
	}

	ss := make(teamStats)

	data := templateData{
		Team:   defe,
		TeamID: teamID,
	}

	for _, gm := range s.Season {
		var gmID int64
		err := db.QueryRowx("SELECT id FROM gamedays WHERE seasonId = $1 AND name = $2", seasonID, gm.Name).Scan(&gmID)
		dieIf(err)

		if gmID == 0 {
			res, err := db.Exec("INSERT INTO gamedays (seasonID, name, date) VALUES ($1, $2, $3)", seasonID, gm.Name, gm.Date)
			dieIf(err)

			gmID, err = res.LastInsertId()
			dieIf(err)
		}

		for _, m := range gm.Matches {
			if m.HomeTeam == "LIBRE" || m.AwayTeam == "LIBRE" {
				continue
			}

			var matchID int64
			err := db.QueryRowx("SELECT id FROM match_results WHERE matchID = $1", m.MatchID).Scan(&matchID)
			dieIf(err)

			if matchID == 0 {
				res, err := db.NamedExec("INSERT INTO match_results (matchId, homeTeam, awayTeam, homeScore, awayScore, date, status) VALUES (:match_id, :home_team, :away_team, :home_score, :away_score, :date, :status)", m)
				dieIf(err)

				matchID, err = res.LastInsertId()
				dieIf(err)
			}

			if m.HomeTeam == defe || m.AwayTeam == defe {
				if !html {
					fmt.Printf("Analizando %s\n", m.Title())
				}

				data.Matches = append(data.Matches, match{m})

				s, err := c.Stats(m)
				dieIf(err)

				var ps []cabb.PlayerStats

				if s.Match.Home == defe {
					ps = s.Stats.Home
				} else {
					ps = s.Stats.Away
				}

				for _, p := range ps {
					if p.PlayedMillis == 0 {
						continue
					}

					s := ss[p.Name]

					s.GamesPlayed += 1

					s.Val += p.Val

					s.Points += p.Points

					s.Shots1P += p.Shots1P
					s.Made1P += p.Made1P

					s.Shots2P += p.Shots2P
					s.Made2P += p.Made2P

					s.Shots3P += p.Shots3P
					s.Made3P += p.Made3P

					s.Assists += p.Assists
					s.Turnovers += p.Turnovers
					s.Steals += p.Steals

					s.Fouls += p.Fouls
					s.Fouled += p.Fouled

					s.Rebounds += p.Rebounds
					s.ReboundsOff += p.ReboundsOff
					s.ReboundsDef += p.ReboundsDef

					s.Blocks += p.Blocks
					s.Blocked += p.Blocked

					s.PlayedMillis += p.PlayedMillis

					ss[p.Name] = s
				}
			}
		}

		if gm.Current {
			break
		}
	}

	data.PlayerStats = ss

	if html {
		dieIf(writeHTML(data))
		return
	}

	ns := make([]string, 0, len(ss))
	for n := range ss {
		ns = append(ns, n)
	}
	sort.Strings(ns)

	fmt.Println()

	var (
		wShots = tabwriter.NewWriter(os.Stdout, 4, 8, 1, ' ', 0)
		wAsTO  = tabwriter.NewWriter(os.Stdout, 4, 8, 1, ' ', 0)
		wFouls = tabwriter.NewWriter(os.Stdout, 4, 8, 1, ' ', 0)
		wRebs  = tabwriter.NewWriter(os.Stdout, 4, 8, 1, ' ', 0)
		wBlks  = tabwriter.NewWriter(os.Stdout, 4, 8, 1, ' ', 0)
		wMins  = tabwriter.NewWriter(os.Stdout, 4, 8, 1, ' ', 0)
	)

	fmt.Fprintln(wShots, "JUGADOR\tPUNTOS\t1P\t2P\t3P")
	fmt.Fprintln(wAsTO, "JUGADOR\tASISTENCIAS\tPÉRDIDAS\tRECUPEROS")
	fmt.Fprintln(wFouls, "JUGADOR\tFOULES\tRECIBIDOS")
	fmt.Fprintln(wRebs, "JUGADOR\tREBOTES\tOFENSIVOS\tDEFENSIVOS")
	fmt.Fprintln(wBlks, "JUGADOR\tTAPONES\tRECIBIDOS")
	fmt.Fprintln(wMins, "JUGADOR\tPARTIDOS\tMINUTOS")

	for _, n := range ns {
		s := ss[n]
		gp := s.GamesPlayed
		ms := time.Millisecond * time.Duration(s.PlayedMillis) / time.Minute

		fmt.Fprintf(wShots, "%s\t%4d (%5.2f)\t%s\t%s\t%s\n",
			n,
			s.Points, float32(s.Points)/float32(gp),
			shots(s.Made1P, s.Shots1P),
			shots(s.Made2P, s.Shots2P),
			shots(s.Made3P, s.Shots3P),
		)

		fmt.Fprintf(wAsTO, "%s\t%3d (%5.2f)\t%3d (%5.2f)\t%3d (%5.2f)\n",
			n,
			s.Assists, float32(s.Assists)/float32(gp),
			s.Turnovers, float32(s.Turnovers)/float32(gp),
			s.Steals, float32(s.Steals)/float32(gp),
		)

		fmt.Fprintf(wFouls, "%s\t%3d (%5.2f)\t%3d (%5.2f)\n",
			n,
			s.Fouls, float32(s.Fouls)/float32(gp),
			s.Fouled, float32(s.Fouled)/float32(gp),
		)

		fmt.Fprintf(wRebs, "%s\t%4d (%5.2f)\t%3d (%5.2f)\t%3d (%5.2f)\n",
			n,
			s.Rebounds, float32(s.Rebounds)/float32(gp),
			s.ReboundsOff, float32(s.ReboundsOff)/float32(gp),
			s.ReboundsDef, float32(s.ReboundsDef)/float32(gp),
		)

		fmt.Fprintf(wBlks, "%s\t%2d (%4.2f)\t%2d (%4.2f)\n",
			n,
			s.Blocks, float32(s.Blocks)/float32(gp),
			s.Blocked, float32(s.Blocked)/float32(gp),
		)

		if n == "TOTALES" {
			continue
		}

		fmt.Fprintf(wMins, "%s\t%2d\t%5.2f\n",
			n,
			gp,
			float32(int(ms))/float32(gp),
		)
	}

	dieIf(wShots.Flush())
	fmt.Println()
	dieIf(wAsTO.Flush())
	fmt.Println()
	dieIf(wFouls.Flush())
	fmt.Println()
	dieIf(wRebs.Flush())
	fmt.Println()
	dieIf(wBlks.Flush())
	fmt.Println()
	dieIf(wMins.Flush())
}

func shots(made, total int) string {
	if total == 0 {
		total = 1
	}
	return fmt.Sprintf("%4d/%4d (%5.2f)", made, total, float64(made)/float64(total))
}

func writeHTML(data templateData) error {
	tpl, err := template.New("").
		Funcs(template.FuncMap{
			"ms": func(ms int64) int { return int(time.Millisecond * time.Duration(ms) / time.Minute) },
			"avg": func(sum, len int) string {
				var avg float32
				if len > 0 {
					avg = float32(sum) / float32(len)
				}
				return fmt.Sprintf("%5.2f", avg)
			},
			"shots": shots,

			"highlight": func(team string, m match) string {
				if (team == m.HomeTeam && m.HomeScore > m.AwayScore) ||
					(team == m.AwayTeam && m.AwayScore > m.HomeScore) {
					return "highlight"
				}
				return ""
			},
			"matchClass": func(m match) string {
				if m.HomeScore > m.AwayScore {
					return "win-home"
				}
				return "win-away"
			},
		}).
		ParseFiles("index.tpl.html")
	if err != nil {
		return fmt.Errorf("parsing HTML template: %w", err)
	}

	return tpl.ExecuteTemplate(os.Stdout, "cabb", data)
}
