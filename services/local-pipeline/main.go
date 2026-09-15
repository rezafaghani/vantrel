package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rezafaghani/vantrel/libs/go/ingestion"
	"github.com/rezafaghani/vantrel/libs/go/marketstore"
	"github.com/rezafaghani/vantrel/libs/go/rawarchive"
	"github.com/rezafaghani/vantrel/libs/go/synthetic"
	"github.com/rezafaghani/vantrel/libs/go/validation"
)

var dashboard = template.Must(template.New("dashboard").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Vantrel Market Pipeline</title>
<style>
body{font-family:system-ui,sans-serif;margin:0;background:#f7f7f4;color:#151515}
main{max-width:1040px;margin:0 auto;padding:32px}
header{display:flex;justify-content:space-between;gap:16px;align-items:center;margin-bottom:24px}
h1{font-size:28px;margin:0}button{padding:10px 14px;border:1px solid #151515;background:#151515;color:white;cursor:pointer}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:12px}
.card{background:white;border:1px solid #ddd;border-radius:6px;padding:16px}
.value{font-size:28px;font-weight:700;margin-top:8px}
pre{white-space:pre-wrap;background:#151515;color:#f7f7f4;padding:16px;border-radius:6px;overflow:auto}
table{width:100%;border-collapse:collapse;background:white;margin:12px 0 24px}
th,td{text-align:left;border-bottom:1px solid #ddd;padding:8px;font-size:14px}
th{font-size:12px;text-transform:uppercase;color:#555}
</style>
</head>
<body><main>
<header><h1>Vantrel Market Pipeline</h1><button id="run">Run Synthetic Batch</button></header>
<section class="grid">
<div class="card">QuestDB<div class="value" id="questdb">...</div></div>
<div class="card">Trades<div class="value" id="trades">...</div></div>
<div class="card">Quotes<div class="value" id="quotes">...</div></div>
<div class="card">Actuals<div class="value" id="actuals">...</div></div>
<div class="card">Forecasts<div class="value" id="forecasts">...</div></div>
</section>
<h2>Last Run</h2><pre id="last">No run yet.</pre>
<h2>Latest Rows</h2><div id="latest">...</div>
</main>
<script>
async function refresh(){
  const r=await fetch('/api/status'); const s=await r.json();
  questdb.textContent=s.questdb_ok?'OK':'Down';
  trades.textContent=s.trades; quotes.textContent=s.quotes; actuals.textContent=s.actuals; forecasts.textContent=s.forecasts;
  const latestResp=await fetch('/api/latest'); renderLatest(await latestResp.json());
}
run.onclick=async()=>{const r=await fetch('/api/run?steps=3',{method:'POST'}); last.textContent=JSON.stringify(await r.json(),null,2); refresh();}
function renderLatest(data){
 latest.innerHTML=['trades','quotes','actuals','forecasts'].map(name=>{
    const rows=data[name]||[];
    if(!rows.length) return '<h3>'+escapeHTML(name)+'</h3><p>No rows.</p>';
    const cols=Object.keys(rows[0]);
    return '<h3>'+escapeHTML(name)+'</h3><table><thead><tr>'+cols.map(c=>'<th>'+escapeHTML(c)+'</th>').join('')+'</tr></thead><tbody>'+rows.map(row=>'<tr>'+cols.map(c=>'<td>'+escapeHTML(row[c]??'')+'</td>').join('')+'</tr>').join('')+'</tbody></table>';
  }).join('');
}
function escapeHTML(value){return String(value).replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));}
refresh();
</script></body></html>`))

func main() {
	steps := flag.Int("steps", 3, "synthetic market steps")
	serve := flag.Bool("serve", false, "serve dashboard and API")
	addr := flag.String("addr", ":8080", "HTTP listen address")
	questdb := flag.String("questdb", "http://127.0.0.1:9000", "QuestDB HTTP base URL")
	schema := flag.String("schema", "../../contracts/questdb/market_observations.sql", "QuestDB schema SQL path")
	flag.Parse()
	if *serve {
		if err := serveApp(*addr, *questdb, *schema); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := run(context.Background(), *steps, time.Now().UTC(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, steps int, now time.Time, out io.Writer) error {
	lines := &strings.Builder{}
	rawBytes, err := runPipeline(ctx, steps, now, marketstore.Publisher{Store: marketstore.NewILPWriter(lines)})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "vantrel local pipeline\nsteps=%d\nraw_archive_bytes=%d\nmarket_ilp_lines=%d\nfirst_ilp_line=%s\n", steps, rawBytes, strings.Count(lines.String(), "\n"), firstLine(lines.String()))
	return err
}

func runPipeline(ctx context.Context, steps int, now time.Time, publisher ingestion.Publisher) (int, error) {
	if steps <= 0 {
		return 0, errors.New("steps must be positive")
	}
	adapter, err := synthetic.Adapter(synthetic.Config{Seed: 1, Steps: steps})
	if err != nil {
		return 0, err
	}
	var raw bytes.Buffer
	sink, err := rawarchive.NewSink(&raw, synthetic.ProviderID, synthetic.SourceID, now)
	if err != nil {
		return 0, err
	}
	err = ingestion.Runtime{
		Adapter: adapter,
		Mode:    ingestion.ModeHistorical,
		Hooks: ingestion.Hooks{
			RawSink:   sink,
			Parser:    synthetic.Parser{},
			Validator: validation.Engine{},
			Publisher: publisher,
		},
	}.Run(ctx)
	if closeErr := sink.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return 0, err
	}
	return raw.Len(), nil
}

func serveApp(addr, questdbURL, schemaPath string) error {
	db := questDB{base: strings.TrimRight(questdbURL, "/"), schemaPath: schemaPath, client: http.DefaultClient}
	if err := db.init(context.Background()); err != nil {
		return err
	}
	return http.ListenAndServe(addr, newMux(db))
}

func newMux(db questDB) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_ = dashboard.Execute(w, nil)
	})
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, db.status(r.Context()))
	})
	mux.HandleFunc("/api/latest", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, db.latest(r.Context()))
	})
	mux.HandleFunc("/api/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		steps, _ := strconv.Atoi(r.URL.Query().Get("steps"))
		if steps == 0 {
			steps = 3
		}
		if err := db.init(r.Context()); err != nil {
			writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		lines := &strings.Builder{}
		rawBytes, err := runPipeline(r.Context(), steps, time.Now().UTC(), marketstore.Publisher{Store: marketstore.NewILPWriter(lines)})
		if err == nil {
			err = db.writeILP(r.Context(), lines.String())
		}
		resp := map[string]any{"ok": err == nil, "steps": steps, "raw_archive_bytes": rawBytes, "market_ilp_lines": strings.Count(lines.String(), "\n")}
		if err != nil {
			resp["error"] = err.Error()
		}
		writeJSON(w, resp)
	})
	return mux
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}
