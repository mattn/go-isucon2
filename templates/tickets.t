{{define "tickets"}}
<h2>{{.Name}}</h2>
<ul>
{{range .Tickets}}
<li class="ticket">
  <a href="/ticket/{{.Id}}">{{.Name}}</a>残り<span class="count">{{.Count}}</span>枚
</li>
{{end}}
</ul>
{{end}}
