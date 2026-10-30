package appserver

import "html/template"

type pageData struct {
	HasFile  bool
	FileName string
	FileSize string
	MaxMB    int64
}

var pageTmpl = template.Must(template.New("index").Parse(pageHTML))

const pageHTML = `<!doctype html>
<html lang="id">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>GoQRShare</title>
<style>
:root{color-scheme:light dark;--bg:#f4f6f9;--card:#fff;--fg:#141922;--muted:#5d6877;--accent:#2563eb;--border:#d8dee6;--ok:#15803d;--err:#b91c1c}
@media (prefers-color-scheme:dark){:root{--bg:#0e1217;--card:#171c24;--fg:#e8ecf2;--muted:#9aa6b5;--accent:#5b9dff;--border:#2a323d;--ok:#4ade80;--err:#f87171}}
*{box-sizing:border-box}
body{margin:0;padding:16px;background:var(--bg);color:var(--fg);font:16px/1.5 system-ui,-apple-system,Segoe UI,Roboto,sans-serif}
main{max-width:520px;margin:0 auto}
h1{font-size:1.3rem;margin:8px 0 16px}
section{background:var(--card);border:1px solid var(--border);border-radius:12px;padding:16px;margin-bottom:16px}
h2{font-size:1rem;margin:0 0 8px}
p{margin:4px 0;color:var(--muted);word-break:break-word}
.btn{display:inline-block;margin-top:10px;padding:12px 18px;border:0;border-radius:10px;background:var(--accent);color:#fff;font-size:1rem;text-decoration:none;cursor:pointer}
.btn:disabled{opacity:.5}
input[type=file]{width:100%;margin-top:6px}
progress{width:100%;margin-top:10px}
#msg{margin-top:10px;word-break:break-word}
.ok{color:var(--ok)}.err{color:var(--err)}
</style>
</head>
<body data-max-mb="{{.MaxMB}}">
<main>
<h1>goqrshare</h1>
{{if .HasFile}}
<section>
<h2>Unduh dari laptop</h2>
<p>{{.FileName}} ({{.FileSize}})</p>
<a class="btn" href="download">Unduh</a>
</section>
{{end}}
<section>
<h2>Kirim ke laptop</h2>
<p>Maksimal {{.MaxMB}} MB per kiriman.</p>
<input id="files" type="file" multiple>
<button id="send" class="btn" type="button">Kirim</button>
<progress id="bar" max="100" value="0" hidden></progress>
<div id="msg" role="status"></div>
</section>
</main>
<script>
(function () {
  var input = document.getElementById('files');
  var btn = document.getElementById('send');
  var bar = document.getElementById('bar');
  var msg = document.getElementById('msg');
  var maxMB = parseInt(document.body.getAttribute('data-max-mb'), 10) || 0;

  function show(text, isErr) {
    msg.textContent = text;
    msg.className = isErr ? 'err' : 'ok';
  }

  btn.addEventListener('click', function () {
    var files = input.files;
    if (!files.length) { show('Pilih file dulu.', true); return; }

    var total = 0, i;
    for (i = 0; i < files.length; i++) { total += files[i].size; }
    if (maxMB && total > maxMB * 1024 * 1024) {
      show('Total ukuran melebihi batas ' + maxMB + ' MB.', true);
      return;
    }

    var fd = new FormData();
    for (i = 0; i < files.length; i++) { fd.append('file', files[i]); }

    var xhr = new XMLHttpRequest();
    xhr.open('POST', 'upload');
    xhr.upload.onprogress = function (e) {
      if (e.lengthComputable) { bar.value = Math.round(e.loaded * 100 / e.total); }
    };
    xhr.onload = function () {
      var r = {};
      try { r = JSON.parse(xhr.responseText); } catch (e) {}
      if (xhr.status === 200) {
        show('Berhasil: ' + (r.filename || ''), false);
        input.value = '';
      } else {
        show(r.msg || ('Gagal (' + xhr.status + ')'), true);
      }
      btn.disabled = false;
    };
    xhr.onerror = function () { show('Koneksi terputus.', true); btn.disabled = false; };

    btn.disabled = true;
    bar.hidden = false;
    bar.value = 0;
    show('Mengirim...', false);
    xhr.send(fd);
  });
})();
</script>
</body>
</html>`
