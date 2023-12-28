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

CREATE TABLE IF NOT EXISTS player_match_results (
       matchId TEXT NOT NULL,
       teamId INTEGER REFERENCES teams (id),
       num INTEGER NOT NULL,
       name TEXT NOT NULL,
       points INTEGER NOT NULL,
       shot1p INTEGER NOT NULL,
       made1p INTEGER NOT NULL,
       missed1p INTEGER NOT NULL,
       assists INTEGER NOT NULL,
       turnovers INTEGER NOT NULL,
       fouls INTEGER NOT NULL,
       fouled INTEGER NOT NULL,
       rebounds INTEGER NOT NULL,
       rebounds_off INTEGER NOT NULL,
       rebounds_def INTEGER NOT NULL,
       played_ms INTEGER NOT NULL,
       played TEXT NOT NULL
 );
