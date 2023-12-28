{{ define "header" }}
<!doctype html>
<html>
  <head>
    <title>{{ .Title }}</title>
    <link rel="stylesheet" href="/static/style.css"/>
  </head>
  <body>
    <h1>{{ .Title }}</h1>
{{ end }}

{{ define "footer" }}
    <script src="/static/cabb.js"></script>
  </body>
</html>
{{ end }}

{{ define "index" }}
{{ template "header" . }}
<table class="teams">
  <colgroup>
    <col/>
    <col class="center"/>
    <col class="center"/>
  </colgroup>
  <thead>
    <tr>
      <th>Equipo</th>
      <th>Temporada</th>
      <th>Estadísticas</th>
    </tr>
  </thead>
  <tbody>
    {{ range .Data }}
    <tr>
      <td>
        <a href="/season/{{ .ID }}">{{ .Name }}</a>
      </td>
      <td>
        <a href="/season/{{ .ID }}" class="button">Ver</a>
      </td>
      <td>
        <a href="/stats/{{ .ID }}" class="button">Ver</a>
      </td>
    </tr>
    {{ end }}
  </tbody>
</table>
{{ template "footer" . }}
{{ end }}

{{ define "season" }}
{{ template "header" . }}
<section id="season">
  <h2>Temporada</h2>
  {{ range .Data.Season }}
  <table class="matches">
    <caption>{{ .Date }} - {{ .Name }}</caption>
    <colgroup>
      <col class="team-name"/>
      <col class="num"/>
      <col class="num"/>
      <col class="team-name"/>
    </colgroup>
    <thead>
      <tr>
        <th colspan="2">Local</th>
        <th colspan="2">Visitante</th>
      </tr>
    </thead>
    <tbody>
      {{ range .Matches }}
      <tr>
        <td>
          <a href="/match/{{ .MatchID }}">{{ .HomeTeam }}</a>
        </td>
        <td>
          <a href="/match/{{ .MatchID }}">{{ .HomeScore }}</a>
        </td>
        <td>
          <a href="/match/{{ .MatchID }}">{{ .AwayScore }}</a>
        </td>
        <td>
          <a href="/match/{{ .MatchID }}">{{ .AwayTeam }}</a>
        </td>
      </tr>
      {{ end }}
    </tbody>
  </table>
  {{ end }}
</section>

<section id="positions">
  <h2>Posiciones</h2>
  <table class="board">
    <thead>
      <tr>
        <th>Nombre</th>
        <th>PJ</th>
        <th>PG</th>
        <th>PP</th>
        <th>PF</th>
        <th>PC</th>
        <th>Puntos</th>
      </tr>
    </thead>
    <tbody>
      {{ range .Data.Positions }}
      <tr>
        <td>{{ .Name }}</td>
        <td>{{ .Played }}</td>
        <td>{{ .Won }}</td>
        <td>{{ .Lost }}</td>
        <td>{{ .Scored }}</td>
        <td>{{ .Received }}</td>
        <td>{{ .Score }}</td>
      </tr>
      {{ end }}
    </tbody>
  </table>
</section>
{{ template "footer" . }}
{{ end }}

{{ define "match" }}
{{ template "header" . }}
<section id="match">
  {{ with .Data.Match }}
  <table id="periods">
    <tr>
      <th>Período</th>
      <th colspan="2">Parcial</th>
    </tr>
    {{ range .Periods }}
    <tr class="num">
      <td>{{ .Period }}</td>
      <td>{{ .HomeScore }}</td>
      <td>{{ .AwayScore }}</td>
    </tr>
    {{ end }}
  </table>
  {{ end }}

  {{ with .Data.Stats }}
  {{ block "players-stats" .Home }}
  <table class="player-stats">
    <colgroup>
      <col class="player-name"/>
      <col span="2" class="num"/>
    </colgroup>
    <thead>
      <th>Nombre</th>
      <th>Mins</th>
      <th>Puntos</th>
    </thead>
    <tbody>
      {{ range . }}{{ if eq .Name "TOTALES" }}{{ continue }}{{ end }}
      {{ block "player-stats" . }}
      <tr>
        <td>{{ .Name }}</td>
        <td>{{ .Played }}</td>
        <td>{{ .Points }}</td>
      </tr>
      {{ end }}
      {{ end }}
    </tbody>
    <tfoot>
      {{ range . }}{{ if eq .Name "TOTALES" }}{{ template "player-stats" . }}{{ end }}
      {{ end }}
    </tfoot>
  </table>
  {{ end }}

  {{ template "players-stats" .Away }}
  {{ end }}
</section>
{{ template "footer" . }}
{{ end }}
