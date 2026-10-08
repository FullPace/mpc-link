// Package web holds the settings page, served by the app on 127.0.0.1 only.
package web

// Page is the whole settings UI: one HTML file, no external resources. It talks to the app's /api/* endpoints with
// the token from its own URL.
const Page = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Padbangers Link</title>
<style>
:root { --bg:#f4f4f2; --card:#ffffff; --ink:#1a1919; --dim:#6b6b6a; --line:#dcdcd8; --accent:#c83d58; --ok:#009797; --warn:#b26a00; }
@media (prefers-color-scheme: dark) { :root { --bg:#161617; --card:#202022; --ink:#ececea; --dim:#9a9a98; --line:#333336; } }
* { box-sizing: border-box; }
body { margin:0; background:var(--bg); color:var(--ink); font:15px/1.45 -apple-system, "Segoe UI", system-ui, sans-serif; }
main { max-width:680px; margin:0 auto; padding:28px 16px 48px; }
h1 { font-size:22px; margin:0 0 4px; } h1 span { color:var(--accent); }
.sub { color:var(--dim); margin:0 0 22px; }
section { background:var(--card); border:1px solid var(--line); border-radius:10px; padding:16px 18px; margin-bottom:16px; }
h2 { font-size:13px; letter-spacing:.06em; text-transform:uppercase; color:var(--dim); margin:0 0 12px; }
label { display:block; font-size:13px; color:var(--dim); margin:10px 0 4px; }
input { width:100%; padding:8px 10px; border:1px solid var(--line); border-radius:7px; background:var(--bg); color:var(--ink); font:inherit; }
.row { display:flex; gap:8px; align-items:center; }
.row input { flex:1; }
button { font:inherit; padding:8px 14px; border-radius:7px; border:1px solid var(--line); background:var(--bg); color:var(--ink); cursor:pointer; white-space:nowrap; }
button.primary { background:var(--accent); border-color:var(--accent); color:#fff; }
button:disabled { opacity:.5; cursor:default; }
.status { display:flex; align-items:center; gap:10px; }
.dot { width:10px; height:10px; border-radius:50%; background:var(--dim); flex:none; }
.dot.ok { background:var(--ok); } .dot.err { background:var(--warn); }
.mount { color:var(--dim); font-size:13px; margin-top:4px; word-break:break-all; }
ul { list-style:none; margin:0; padding:0; }
li { display:flex; gap:8px; align-items:center; padding:8px 0; border-top:1px solid var(--line); }
li:first-child { border-top:0; }
li .name { width:34%; }
li .path { flex:1; color:var(--dim); font-size:13px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.empty { color:var(--dim); font-size:14px; padding:4px 0 8px; }
.found button { margin:6px 6px 0 0; }
.hint { color:var(--dim); font-size:13px; margin:10px 0 0; }
</style>
</head>
<body>
<main>
  <h1><span>Padbangers</span> Link</h1>
  <p class="sub">Your computer's folders in the MPC's browser, over the network.</p>

  <section>
    <h2>Status</h2>
    <div class="status"><div id="dot" class="dot"></div><div><div id="statusText">…</div><div id="mount" class="mount"></div></div></div>
    <p class="hint"><button id="toggle">Disconnect</button></p>
  </section>

  <section>
    <h2>MPC</h2>
    <label for="host">Address</label>
    <div class="row"><input id="host" placeholder="192.168.1.20"><button id="find">Find MPC</button></div>
    <div id="found" class="found"></div>
    <label for="cname">Folder name on the MPC</label>
    <input id="cname">
    <p class="hint">The MPC shows your folders under <b>MPC SSD</b> (or MPC Documents) in a folder with this name.</p>
    <p class="hint"><button class="primary" id="save">Save</button></p>
  </section>

  <section>
    <h2>Shared folders</h2>
    <ul id="folders"></ul>
    <p class="hint"><button class="primary" id="add">Add folder…</button></p>
  </section>

  <p class="hint" id="ver"></p>
</main>
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
    (!s.host ? "Set the MPC's address below" : (s.folders.length ? (st.err ? "Not connected: " + st.err : "Connecting…") : "Add a folder to share"));
  $("mount").textContent = st.connected ? "On the MPC: " + st.mount.replace("/media/", "") : "";
  $("toggle").textContent = s.enabled ? "Disconnect" : "Connect";
  if (!editing) { $("host").value = s.host || ""; $("cname").value = s.computer_name || ""; $("cname").placeholder = s.name; }
  const ul = $("folders"); ul.innerHTML = "";
  if (!s.folders.length) ul.innerHTML = '<div class="empty">No folders yet.</div>';
  s.folders.forEach((f, i) => {
    const li = document.createElement("li");
    const name = document.createElement("input"); name.className = "name"; name.value = f.name;
    name.onchange = () => call("/api/folders/rename?i=" + i + "&name=" + encodeURIComponent(name.value)).then(render);
    const path = document.createElement("div"); path.className = "path"; path.textContent = f.path; path.title = f.path;
    const rm = document.createElement("button"); rm.textContent = "Remove";
    rm.onclick = () => call("/api/folders/remove?i=" + i).then(render);
    li.append(name, path, rm); ul.append(li);
  });
  $("ver").textContent = "Padbangers Link " + s.version;
}
["host", "cname"].forEach(id => { $(id).onfocus = () => editing = true; });
$("save").onclick = async () => { editing = false; render(await call("/api/settings", { host: $("host").value.trim(), computer_name: $("cname").value.trim() })); };
$("add").onclick = async () => { $("add").disabled = true; try { render(await call("/api/folders/add")); } finally { $("add").disabled = false; } };
$("toggle").onclick = async () => render(await call($("toggle").textContent === "Connect" ? "/api/connect" : "/api/disconnect"));
$("find").onclick = async () => {
  $("find").disabled = true; $("found").textContent = "Searching the network…";
  try {
    const list = await call("/api/discover");
    $("found").innerHTML = list.length ? "" : "No MPC found. Is it on the same network?";
    list.forEach(m => { const b = document.createElement("button"); b.textContent = m.name + " (" + m.host + ")";
      b.onclick = () => { $("host").value = m.host; editing = true; }; $("found").append(b); });
  } catch (e) { $("found").textContent = e.message; }
  $("find").disabled = false;
};
async function poll() { try { render(await call("/api/state")); } catch (e) {} setTimeout(poll, 2000); }
poll();
</script>
</body>
</html>`
