// Package web holds the settings page, served by the app on 127.0.0.1 only.
package web

// Page is the whole settings UI in padbangers.com's look (its global.css tokens: cream panels, dark header with the
// banner, cyan buttons, never red ones). One HTML file; the banner comes from the app (/banner.png). It talks to the
// app's /api/* endpoints with the token from its own URL.
const Page = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>MPC Link</title>
<style>
:root {
  --bg:#ecebe6; --bg-2:#e3e1da; --panel:#f4f3ef; --edge:#c4c2ba; --edge-dark:#a6a49c;
  --display:#161616; --display-2:#232323; --text:#2a2a28; --text-dim:#5f6061;
  --cyan:#1ec8e6; --cyan-dim:#15a6c0; --orange:#ff7a00; --orange-dim:#e86a00; --red:#e30613;
  --radius:12px; --radius-sm:8px; --transition:0.18s ease;
}
* { box-sizing:border-box; }
body { margin:0; background:var(--bg); color:var(--text); font-family:"Helvetica Neue", Arial, sans-serif;
  font-size:16px; line-height:1.6; -webkit-font-smoothing:antialiased; min-height:100vh; display:flex; flex-direction:column; }
main { flex:1 0 auto; }
:focus-visible { outline:2px solid var(--cyan); outline-offset:3px; }

.site-header { background:linear-gradient(var(--display-2), var(--display)); border-bottom:2px solid #000; }
.site-header__inner { display:flex; align-items:center; justify-content:space-between; gap:22px; max-width:820px;
  margin:0 auto; padding:18px 28px; }
.site-header__logo { width:clamp(180px, 36vw, 260px); height:auto; display:block; }
.app-name { color:#fff; font-size:20px; font-weight:700; letter-spacing:.02em; white-space:nowrap; }
.app-name span { color:var(--cyan); }

.page { max-width:820px; margin-inline:auto; padding:36px 28px 64px; width:100%; }
.page h1 { font-size:clamp(26px, 5vw, 34px); margin:0 0 6px; line-height:1.2; }
.page__lead { color:var(--text-dim); font-size:17px; margin:0 0 28px; }

.card { background:var(--panel); border:1px solid var(--edge); border-radius:var(--radius); padding:22px 24px; margin-bottom:22px; }
.card h2 { font-size:20px; margin:0 0 14px; padding-bottom:8px; border-bottom:1px solid var(--edge); line-height:1.2; }

.field { display:grid; gap:6px; margin-bottom:16px; }
.field label { font-size:14px; font-weight:600; }
input { font:inherit; color:var(--text); background:var(--bg); border:1px solid var(--edge-dark); border-radius:var(--radius-sm);
  padding:10px 12px; width:100%; transition:border-color var(--transition); }
input:focus { outline:none; border-color:var(--cyan-dim); }
.row { display:flex; gap:10px; align-items:center; }
.row input { flex:1; }
.hint { color:var(--text-dim); font-size:14px; margin:4px 0 0; }

.btn { display:inline-flex; align-items:center; gap:8px; border-radius:var(--radius-sm); padding:11px 22px; font:inherit;
  font-weight:700; font-size:15px; cursor:pointer; border:1px solid transparent; white-space:nowrap;
  transition:filter var(--transition), background var(--transition), border-color var(--transition); }
.btn--primary { background:linear-gradient(var(--cyan), var(--cyan-dim)); color:#fff; }
.btn--primary:hover { filter:brightness(1.08); }
.btn--ghost { background:transparent; color:var(--text); border-color:var(--edge-dark); }
.btn--ghost:hover { background:var(--bg-2); }
.btn--small { padding:7px 14px; font-size:14px; }
.btn:disabled { opacity:.5; cursor:default; filter:none; }
.actions { display:flex; gap:10px; flex-wrap:wrap; margin-top:6px; }

.status { display:flex; align-items:center; gap:14px; }
.dot { width:14px; height:14px; border-radius:50%; background:var(--edge-dark); flex:none; }
.dot.ok { background:var(--cyan); box-shadow:0 0 0 4px rgba(30,200,230,.18); }
.dot.err { background:var(--orange); box-shadow:0 0 0 4px rgba(255,122,0,.18); }
.status__text { font-weight:700; }
.status__mount { color:var(--text-dim); font-size:14px; word-break:break-all; }

ul { list-style:none; margin:0 0 6px; padding:0; }
li { display:flex; gap:10px; align-items:center; padding:10px 0; border-top:1px solid var(--edge); }
li:first-child { border-top:0; padding-top:0; }
li input { width:34%; flex:none; padding:7px 10px; }
li .path { flex:1; color:var(--text-dim); font-size:14px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.empty { color:var(--text-dim); font-size:15px; padding:0 0 10px; }
.found { display:flex; flex-wrap:wrap; gap:8px; margin-top:10px; color:var(--text-dim); font-size:14px; }

.site-footer { flex-shrink:0; background:linear-gradient(var(--display-2), var(--display)); color:#cfcec9;
  text-align:center; padding:22px 28px; font-size:13px; }
.site-footer a { color:var(--cyan); text-decoration:none; }
.site-footer small { color:#8b8c8d; display:block; margin-top:6px; }

@media (max-width:600px) {
  .site-header__inner { flex-direction:column; gap:10px; }
  li { flex-wrap:wrap; } li input { width:100%; }
}
</style>
</head>
<body>
<header class="site-header">
  <div class="site-header__inner">
    <img class="site-header__logo" src="/banner.png" alt="Padbangers">
    <div class="app-name">MPC <span>Link</span></div>
  </div>
</header>
<main>
  <div class="page">
    <h1>MPC Link</h1>
    <p class="page__lead">Your computer's folders in your MPC's browser, over the network. No controller mode, no copying.</p>

    <section class="card">
      <h2>Status</h2>
      <div class="status">
        <div id="dot" class="dot"></div>
        <div><div id="statusText" class="status__text">…</div><div id="mount" class="status__mount"></div></div>
      </div>
      <div class="actions" style="margin-top:16px"><button class="btn btn--ghost btn--small" id="toggle">Disconnect</button></div>
    </section>

    <section class="card">
      <h2>Your MPC</h2>
      <div class="field">
        <label for="host">Address</label>
        <div class="row"><input id="host" placeholder="192.168.1.20" autocomplete="off"><button class="btn btn--ghost" id="find">Find MPC</button></div>
        <div id="found" class="found"></div>
      </div>
      <div class="field">
        <label for="cname">Folder name on the MPC</label>
        <input id="cname" autocomplete="off">
        <p class="hint">Your folders show up under <b>MPC SSD</b> (or MPC Documents) in a folder with this name.</p>
      </div>
      <div class="actions"><button class="btn btn--primary" id="save">Save</button></div>
    </section>

    <section class="card">
      <h2>Shared folders</h2>
      <ul id="folders"></ul>
      <div class="actions"><button class="btn btn--primary" id="add">Add folder…</button></div>
    </section>
  </div>
</main>
<footer class="site-footer">
  <a href="https://padbangers.com" target="_blank" rel="noopener">padbangers.com</a>
  <small id="ver"></small>
</footer>
<script>
const T = new URLSearchParams(location.search).get("t");
const $ = id => document.getElementById(id);
async function call(path, body) {
  const r = await fetch(path, { method: "POST", headers: { "X-Token": T, "Content-Type": "application/json" }, body: body ? JSON.stringify(body) : undefined });
  const j = await r.json();
  if (!r.ok) throw new Error(j.error || r.statusText);
  return j;
}
let editing = false;
function render(s) {
  const st = s.status || {};
  $("dot").className = "dot" + (st.connected ? " ok" : (s.enabled && st.err ? " err" : ""));
  $("statusText").textContent = !s.enabled ? "Disconnected" : st.connected ? "Connected" :
    (!s.host ? "Set your MPC's address below" : (s.folders.length ? (st.err ? "Not connected: " + st.err : "Connecting…") : "Add a folder to share"));
  $("mount").textContent = st.connected ? "On the MPC: " + st.mount.replace("/media/", "").replace("/", " / ") : "";
  $("toggle").textContent = s.enabled ? "Disconnect" : "Connect";
  if (!editing) { $("host").value = s.host || ""; $("cname").value = s.computer_name || ""; $("cname").placeholder = s.name; }
  const ul = $("folders"); ul.innerHTML = "";
  if (!s.folders.length) ul.innerHTML = '<div class="empty">No folders yet. Add the folders you want on your MPC.</div>';
  s.folders.forEach((f, i) => {
    const li = document.createElement("li");
    const name = document.createElement("input"); name.value = f.name; name.title = "Name on the MPC";
    name.onchange = () => call("/api/folders/rename?i=" + i + "&name=" + encodeURIComponent(name.value)).then(render);
    const path = document.createElement("div"); path.className = "path"; path.textContent = f.path; path.title = f.path;
    const rm = document.createElement("button"); rm.className = "btn btn--ghost btn--small"; rm.textContent = "Remove";
    rm.onclick = () => call("/api/folders/remove?i=" + i).then(render);
    li.append(name, path, rm); ul.append(li);
  });
  $("ver").textContent = "MPC Link " + s.version + " · © Padbangers";
}
["host", "cname"].forEach(id => { $(id).onfocus = () => editing = true; });
$("save").onclick = async () => { editing = false; render(await call("/api/settings", { host: $("host").value.trim(), computer_name: $("cname").value.trim() })); };
$("add").onclick = async () => { $("add").disabled = true; try { render(await call("/api/folders/add")); } finally { $("add").disabled = false; } };
$("toggle").onclick = async () => render(await call($("toggle").textContent === "Connect" ? "/api/connect" : "/api/disconnect"));
$("find").onclick = async () => {
  $("find").disabled = true; $("found").textContent = "Searching your network…";
  try {
    const list = await call("/api/discover");
    $("found").innerHTML = list.length ? "" : "No MPC found. Is it switched on and on the same network?";
    list.forEach(m => { const b = document.createElement("button"); b.className = "btn btn--ghost btn--small";
      b.textContent = m.name + " · " + m.host; b.onclick = () => { $("host").value = m.host; editing = true; }; $("found").append(b); });
  } catch (e) { $("found").textContent = e.message; }
  $("find").disabled = false;
};
async function poll() { try { render(await call("/api/state")); } catch (e) {} setTimeout(poll, 2000); }
poll();
</script>
</body>
</html>`
