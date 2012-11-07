{{define "seats"}}
<h2>{{.Ticket.ArtistName}} : {{.Ticket.Name}}</h2>
<ul>
{{range .Variations}}
<li class="variation">
  <form method="POST" action="/buy">
    <input type="hidden" name="ticket_id" value="{{.TicketId}}">
    <input type="hidden" name="variation_id" value="{{.Id}}">
    <span class="variation_name">{{.Name}}</span> 残り<span class="vacancy" id="{{.Id}}">{{.Vacancy}}</span>席
    <input type="text" name="member_id" value="">
    <input type="submit" value="購入">
  </form>
</li>
{{end}}
</ul>

<h3>席状況</h3>
{{range .Variations}}
<h4>{{.Name}}</h4>
<table class="seats" data-variationid="{{.Id}}">
{{range .SeatIds}}
  <tr>
{{range .}}
    <td id="" class="{{if .}}unavailable{{else}}available{{end}}"></td>
{{end}}
  </tr>
{{end}}
</table>
{{end}}
{{end}}
