{{define "recent"}}
      <table>
        <tr><th colspan="2">最近購入されたチケット</th></tr>
{{range .RecentSolds}}
        <tr>
          <td class="recent_variation">{{.VName}} {{.TName}} {{.AName}}</td>
          <td class="recent_seat_id">{{.Id}}</td>
        </tr>
{{end}}
      </table>
{{end}}
