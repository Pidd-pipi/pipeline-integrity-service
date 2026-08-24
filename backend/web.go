package main

import (
	"net/http"
	"strings"
)

func staticHandler() http.Handler {
	const index = `<!doctype html><html><head><meta charset="utf-8"><title>Pipeline integrity</title></head><body><h1>Pipeline integrity cycles</h1><p id="health">Loading...</p><p id="latency">Latency samples: <!--latency--></p><button id="refresh">Refresh cycles</button><ul id="items"></ul><script src="/app.js"></script></body></html>`
	const app = `const list=document.querySelector('#items');async function load(){const xs=await (await fetch('/api/cycles')).json();list.innerHTML=xs.map(x=>'<li>'+x.segment+' / '+x.method+' - '+x.status+' <button data-id="'+x.id+'">Accept</button></li>').join('');document.querySelectorAll('[data-id]').forEach(b=>b.onclick=async()=>{await fetch('/api/cycles/'+b.dataset.id+'/status',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({status:'accepted'})});load()})}fetch('/healthz').then(r=>r.json()).then(x=>document.querySelector('#health').textContent=x.status+' / '+x.service);document.querySelector('#refresh').onclick=load;load();`
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			page := strings.Replace(index, "<!--latency-->", formatOpsInt(len(opsPathLatency["/api/cycles"])), 1)
			_, _ = w.Write([]byte(page))
			return
		}
		if r.URL.Path == "/app.js" {
			w.Header().Set("Content-Type", "application/javascript")
			_, _ = w.Write([]byte(app))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/web/") {
			http.NotFound(w, r)
			return
		}
		http.NotFound(w, r)
	})
}
