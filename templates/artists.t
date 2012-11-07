{{define "artists"}}
<h1>TOP</h1>
<ul>
{{range .Artists}}
  <li>
    <a href="/artist/{{.Id}}">
      <span class="artist_name">{{.Name}}</span>
    </a>
  </li>
{{end}}
</ul>
{{end}}
