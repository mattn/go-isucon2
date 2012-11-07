{{define "admin"}}<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <title>isucon 2</title>
    <link type="text/css" rel="stylesheet" href="/css/ui-lightness/jquery-ui-1.8.24.custom.css">
    <link type="text/css" rel="stylesheet" href="/css/isucon2.css">
    <script type="text/javascript" src="/js/jquery-1.8.2.min.js"></script>
    <script type="text/javascript" src="/js/jquery-ui-1.8.24.custom.min.js"></script>
    <script type="text/javascript" src="/js/isucon2.js"></script>
  </head>
  <body>
    <header>
      <a href="/">
        <img src="/images/isucon_title.jpg">
      </a>
    </header>
    <div id="sidebar">
    </div>
    <div id="content">
    <ul>
      <li>
        <a href="/admin/order.csv">注文csv</a>
      </li>
      <li>
        <form method="post">
          <input type="submit" value="データ初期化">
        </form>
      </li>
    </ul>
    </div>
  </body>
</html>
{{end}}
