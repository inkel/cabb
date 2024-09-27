package cabb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	_ "github.com/mattn/go-sqlite3"
)

type Client struct {
	deviceID string
	key      string
	db       *sql.DB
}

func NewClient(uid, deviceID string) (Client, error) {
	data := url.Values{
		"uid":              {uid},
		"plataforma":       {"ios"},
		"tipo_dispositivo": {"mobile"},
		"token_push":       {},
		"version":          {"30012"},
		"accion":           {"acceso"},
	}

	c := Client{
		deviceID: deviceID,
	}

	if D {
		db, err := sql.Open("sqlite3", "requests.db")
		if err != nil {
			return c, err
		}

		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS requests (url TEXT, qs TEXT, body JSON)`)
		if err != nil {
			return c, err
		}

		c.db = db
	}

	var r struct {
		cabbResponseGeneric
		Key string `json:"key"`
	}

	if err := c.request("dispositivo.ashx", data, &r); err != nil {
		return c, fmt.Errorf("initiating connection: %w", err)
	}

	c.key = r.Key

	return c, nil
}

const baseURL = "https://appaficioncabb.indalweb.net/"

type cabbResponse interface {
	CABBError() error
}

type cabbResponseGeneric struct {
	Result string `json:"resultado"`
	Error  string `json:"error"`
}

func (r cabbResponseGeneric) CABBError() error {
	if r.Result != "error" {
		return nil
	}
	return fmt.Errorf("error response: %s", r.Error)
}

var D bool

func (c Client) request(path string, data url.Values, d cabbResponse) error {
	if data == nil {
		data = url.Values{}
	}

	url, err := url.JoinPath(baseURL, path)
	if err != nil {
		return fmt.Errorf("building URL for path %s: %w", path, err)
	}

	data.Set("id_dispositivo", c.deviceID)

	if c.key != "" {
		data.Set("key", c.key)
	}

	res, err := http.PostForm(url, data)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if D {
		c.db.Exec(`INSERT INTO requests (url, qs, body) VALUES (?, ?, ?)`, url, data.Encode(), body)
	}

	if err := json.Unmarshal(body, &d); err != nil {
		return err
	}

	if err := d.CABBError(); err != nil {
		return fmt.Errorf("error: %s", err)
	}

	return nil
}

type Team struct {
	NotificationID string `json:"idEquipoNotificacion"`
	ID             string `json:"id"`
	Club           string `json:"club"`
	Name           string `json:"nombre"`
}

func (c Client) Teams() ([]Team, error) {
	var r struct {
		cabbResponseGeneric
		Teams []Team `json:"misequipos"`
	}

	if err := c.request("misequiposV2.ashx", url.Values{"accion": {"listado"}}, &r); err != nil {
		return nil, fmt.Errorf("fetching teams: %w", err)
	}

	return r.Teams, nil
}

type GameDay struct {
	Name    string  `json:"jornada"`
	Date    string  `json:"fecha"`
	Current bool    `json:"activa"`
	Matches []Match `json:"partidos"`
}

type Match struct {
	MatchID   string `json:"idPartido" db:"match_id"`
	HomeTeam  string `json:"nombreEquipo1" db:"home_team"`
	AwayTeam  string `json:"nombreEquipo2" db:"away_team"`
	HomeScore string `json:"puntosEquipo1" db:"home_score"`
	AwayScore string `json:"puntosEquipo2" db:"away_score"`
	Date      string `json:"fecha" db:"date"`
	Time      string `json:"hora" db:"time"`
	Status    string `json:"estado" db:"status"`
}

func (m Match) Title() string {
	return fmt.Sprintf("%s %s - %s %s", m.HomeTeam, m.HomeScore, m.AwayTeam, m.AwayScore)
}

type Position struct {
	Name     string `json:"nombre"`
	Pos      int    `json:"posicion"`
	Played   int    `json:"pj"`
	Won      int    `json:"pg"`
	Lost     int    `json:"pp"`
	ID       int    `json:"id"` // TODO same as Team.NotificationID?
	Score    int    `json:"puntos"`
	Scored   int    `json:"pf"`
	Received int    `json:"pc"`
}

type Season struct {
	cabbResponseGeneric
	TeamID    string
	Season    []GameDay  `json:"jornadas"`
	Positions []Position `json:"clasificacion"`
}

func (c Client) Season(teamID string) (Season, error) {
	s := Season{TeamID: teamID}

	data := url.Values{
		"accion":    {"detalleEquipo"},
		"id_equipo": {teamID},
	}

	if err := c.request("misequiposV2.ashx", data, &s); err != nil {
		return s, err
	}

	return s, nil
}

type Stats struct {
	cabbResponseGeneric
	Match   LiveMatch `json:"partido"`
	MatchID string
	Stats   struct {
		Home []PlayerStats `json:"estadisticasequipolocal"`
		Away []PlayerStats `json:"estadisticasequipovisitante"`
	} `json:"estadisticas"`
}

type PlayerStats struct {
	Num  string `json:"dorsal" db:"num"`
	Name string `json:"nombre" db:"name"`
	Val  int    `json:"valoracion" db:"val"`

	Points int `json:"puntos" db:"points"`

	Shots1P  int `json:"tiro1p" db:"shot1p"`
	Made1P   int `json:"canasta1p" db:"made1p"`
	Missed1p int `json:"tiro1pFallado" db:"missed1p"`

	Shots2P  int `json:"tiro2p" db:"shot2p"`
	Made2P   int `json:"canasta2p" db:"made2p"`
	Missed2p int `json:"tiro2pFallado" db:"missed2p"`

	Shots3P  int `json:"tiro3p" db:"shot3p"`
	Made3P   int `json:"canasta3p" db:"made3p"`
	Missed3p int `json:"tiro3pFallado" db:"missed3p"`

	Assists   int `json:"asistencias" db:"assists"`
	Turnovers int `json:"perdidas" db:"turnovers"`
	Steals    int `json:"recuperaciones" db:"steals"`

	Fouls  int `json:"faltascometidas" db:"fouls"`
	Fouled int `json:"faltasrecibidas" db:"fouled"`

	Rebounds    int `json:"rebotes" db:"rebounds"`
	ReboundsOff int `json:"reboteofensivo" db:"rebounds_off"`
	ReboundsDef int `json:"rebotedefensivo" db:"rebounds_def"`

	Blocks  int `json:"taponescometidos"`
	Blocked int `json:"taponesrecibidos"`

	PlayedMillis int64  `json:"milisegundos_jugados" db:"played_ms"`
	Played       string `json:"tiempo_jugado" db:"played"`
}

func (c Client) Stats(m Match) (Stats, error) {
	var s = Stats{MatchID: m.MatchID}

	if err := c.request("envivo/estadisticas.ashx", url.Values{"id_partido": {m.MatchID}}, &s); err != nil {
		return s, err
	}

	return s, nil
}

type Period struct {
	Period    int `json:"periodo"`
	HomeScore int `json:"tanteo_periodo_local"`
	AwayScore int `json:"tanteo_periodo_visitante"`
}

type LiveMatch struct {
	Home       string   `json:"local"`
	HomeID     int      `json:"idlocal"`
	HomeScore  int      `json:"tanteo_local"`
	Away       string   `json:"visitante"`
	AwayID     int      `json:"idvisitante"`
	AwayScore  int      `json:"tanteo_visitante"`
	NumPeriods int      `json:"numperiodos"`
	Overtime   bool     `json:"tiene_prorrogas"`
	Periods    []Period `json:"periodos"`
}

type Action struct {
	ActionNum int    `json:"autoincremental_id"`
	Type      string `json:"accion_tipo"`
	Info      string `json:"informacion_adicional"`
	Period    int    `json:"numero_periodo"`
	MatchTime string `json:"tiempo_partido"`
	TeamID    int    `json:"equipo_id"`
	PlayerNum string `json:"dorsal"`
	ActorID   string `json:"componente_id"`
}

type Live struct {
	cabbResponseGeneric
	LiveMatch LiveMatch `json:"partido"`
	Match     Match
	Live      struct {
		Actions []Action `json:"historialacciones"`
	} `json:"envivo"`
}

func (c Client) Live(m Match) (Live, error) {
	var l Live

	//D = true
	if err := c.request("envivo/partido.ashx", url.Values{"id_partido": {m.MatchID}}, &l); err != nil {
		return l, nil
	}
	l.Match = m
	//D = false

	return l, nil
}

type Leagues struct {
	cabbResponseGeneric
	Leagues []League `json:"delegaciones"`
}

type League struct {
	Name string `json:"provincia"` // JFC
}

func (c Client) Leagues() ([]League, error) {
	var r Leagues
	if err := c.request("delegaciones.ashx", nil, &r); err != nil {
		return nil, err
	}
	return r.Leagues, nil
}

type Tournament struct {
	ID   string `json:"id"`
	Name string `json:"nombre"`
}

type Tournaments struct {
	cabbResponseGeneric
	Tournaments []Tournament `json:"valores"`
}

func (c Client) Tournaments(league string) ([]Tournament, error) {
	var r Tournaments
	if err := c.request("equipos-jugadores.ashx", url.Values{"accion": {"competiciones"}, "delegacion": {league}}, &r); err != nil {
		return nil, err
	}
	return r.Tournaments, nil
}

type Categories struct {
	cabbResponseGeneric
	Categories []Category `json:"valores"`
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"nombre"`
}

func (c Client) Categories(t Tournament) ([]Category, error) {
	var r Categories
	if err := c.request("equipos-jugadores.ashx", url.Values{"accion": {"categorias"}, "competicion": {t.ID}}, &r); err != nil {
		return nil, err
	}
	return r.Categories, nil
}

type Clubs struct {
	cabbResponseGeneric
	Clubs []Club `json:"valores"`
}

type Club struct {
	ID   string `json:"id"`
	Name string `json:"nombre"`
}

func (c Client) Clubs(t Tournament, cat Category) ([]Club, error) {
	var r Clubs
	if err := c.request("equipos-jugadores.ashx", url.Values{"accion": {"clubes"}, "categoria": {cat.ID}, "competicion": {t.ID}}, &r); err != nil {
		return nil, err
	}
	return r.Clubs, nil
}

type Teams struct {
	ID   string `json:"id"`
	Name string `json:"nombre"`
}
