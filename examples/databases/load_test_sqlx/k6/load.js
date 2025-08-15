import http from "k6/http";
import { check, sleep, group } from "k6";
import { Trend } from "k6/metrics";

// ===== Custom Trends (MILLISECONDS) =====
const loop_stmt_select = new Trend("loop_stmt_select");
const loop_nonstmt_select = new Trend("loop_nonstmt_select");
const loop_stmt_insert = new Trend("loop_stmt_insert");
const loop_nonstmt_insert = new Trend("loop_nonstmt_insert");

const loop_stmt_select_heavy = new Trend("loop_stmt_select_heavy");
const loop_nonstmt_select_heavy = new Trend("loop_nonstmt_select_heavy");
const loop_stmt_insert_heavy = new Trend("loop_stmt_insert_heavy");
const loop_nonstmt_insert_heavy = new Trend("loop_nonstmt_insert_heavy");

// ===== Options =====
export const options = {
  thresholds: {
    // global http
    http_req_duration: ["p(95)<1500"],

    // light endpoints
    "http_req_duration{endpoint:stmt_select}": ["p(95)<1200"],
    "http_req_duration{endpoint:stmt_insert}": ["p(95)<2000"],
    "http_req_duration{endpoint:nonstmt_select}": ["p(95)<1000"],
    "http_req_duration{endpoint:nonstmt_insert}": ["p(95)<1100"],

    // heavy endpoints (longer SLA)
    "http_req_duration{endpoint:stmt_select_heavy}": ["p(95)<2500"],
    "http_req_duration{endpoint:nonstmt_select_heavy}": ["p(95)<3000"],
    "http_req_duration{endpoint:stmt_insert_heavy}": ["p(95)<2500"],
    "http_req_duration{endpoint:nonstmt_insert_heavy}": ["p(95)<3000"],

    // per-loop metrics (server-reported avg_per_loop → ms)
    loop_stmt_select: ["p(95)<40"],
    loop_nonstmt_select: ["p(95)<60"],
    loop_stmt_insert: ["p(95)<45"],
    loop_nonstmt_insert: ["p(95)<65"],

    // heavy per-loop (lebih longgar)
    loop_stmt_select_heavy: ["p(95)<400"],
    loop_nonstmt_select_heavy: ["p(95)<550"],
    loop_stmt_insert_heavy: ["p(95)<450"],
    loop_nonstmt_insert_heavy: ["p(95)<600"],
  },

  // ramping VU
  stages: [
    { duration: "15s", target: 20 },
    { duration: "45s", target: 60 },
    { duration: "60s", target: 120 },
    { duration: "20s", target: 0 },
  ],
};

const BASE = __ENV.TARGET || "http://host.docker.internal:8080";

// ===== Heavy params pool (randomized each iteration) =====
const HEAVY_Q = (
  __ENV.HEAVY_Q || "product,features,great,awesome sale,new popular"
)
  .split(/[,\s]+/)
  .filter(Boolean);
const HEAVY_BRANDS = (
  __ENV.HEAVY_BRANDS || "brand5,brand11,brand17,brand23,brand31,brand42"
)
  .split(",")
  .map((s) => s.trim())
  .filter(Boolean);
const HEAVY_TOPK_MIN = Number(__ENV.HEAVY_TOPK_MIN || 5);
const HEAVY_TOPK_MAX = Number(__ENV.HEAVY_TOPK_MAX || 15);
const HEAVY_LIMIT_MIN = Number(__ENV.HEAVY_LIMIT_MIN || 100);
const HEAVY_LIMIT_MAX = Number(__ENV.HEAVY_LIMIT_MAX || 300);

// ===== Warm-up =====
export function setup() {
  http.get(`${BASE}/stmt/select?name=bench-user&loops=50&tx=false`, {
    tags: { endpoint: "warmup" },
  });
  http.get(`${BASE}/stmt/insert?prefix=warm&loops=50&tx=false&keep=false`, {
    tags: { endpoint: "warmup" },
  });

  // heavy warmup (kecil)
  const q = pickOne(HEAVY_Q);
  const brand = pickOne(HEAVY_BRANDS);
  http.get(
    `${BASE}/stmt/select/heavy?q=${encodeURIComponent(
      q
    )}&brand=${encodeURIComponent(brand)}&top_per_category=5&limit=80&offset=0`,
    { tags: { endpoint: "warmup" } }
  );
  http.get(
    `${BASE}/stmt/insert/heavy?orders=50&items_per_order=2&q=${encodeURIComponent(
      q
    )}&brand=${encodeURIComponent(brand)}&keep=false`,
    { tags: { endpoint: "warmup" } }
  );
}

// ===== Main =====
export default function () {
  const loops = 20 + Math.floor(Math.random() * 10);
  const name = "bench-user";

  // randomized heavy params each iteration
  const q = pickOne(HEAVY_Q);
  const brand = pickOne(HEAVY_BRANDS);
  const topK = randInt(HEAVY_TOPK_MIN, HEAVY_TOPK_MAX);
  const limit = randInt(HEAVY_LIMIT_MIN, HEAVY_LIMIT_MAX);
  const offset = 0;
  const orders = 150 + Math.floor(Math.random() * 150); // heavy insert volume
  const itemsPerOrder = 2 + Math.floor(Math.random() * 3); // 2..4 items each

  group("select_light", () => {
    recordSelect(
      `${BASE}/stmt/select?name=${name}&loops=${loops}&tx=true`,
      "stmt_select",
      loop_stmt_select
    );
    recordSelect(
      `${BASE}/stmt/select?name=${name}&loops=${loops}&tx=false`,
      "stmt_select",
      loop_stmt_select
    );
    recordSelect(
      `${BASE}/nonstmt/select?name=${name}&loops=${loops}&tx=true`,
      "nonstmt_select",
      loop_nonstmt_select
    );
    recordSelect(
      `${BASE}/nonstmt/select?name=${name}&loops=${loops}&tx=false`,
      "nonstmt_select",
      loop_nonstmt_select
    );
  });

  group("insert_light", () => {
    recordInsert(
      `${BASE}/stmt/insert?prefix=bench-ins&loops=${loops}&tx=true&keep=false`,
      "stmt_insert",
      loop_stmt_insert
    );
    recordInsert(
      `${BASE}/stmt/insert?prefix=bench-ins&loops=${loops}&tx=false&keep=false`,
      "stmt_insert",
      loop_stmt_insert
    );
    recordInsert(
      `${BASE}/nonstmt/insert?prefix=bench-ins&loops=${loops}&tx=true&keep=false`,
      "nonstmt_insert",
      loop_nonstmt_insert
    );
    recordInsert(
      `${BASE}/nonstmt/insert?prefix=bench-ins&loops=${loops}&tx=false&keep=false`,
      "nonstmt_insert",
      loop_nonstmt_insert
    );
  });

  group("select_heavy", () => {
    recordSelect(
      `${BASE}/stmt/select/heavy?q=${encodeURIComponent(
        q
      )}&brand=${encodeURIComponent(
        brand
      )}&top_per_category=${topK}&limit=${limit}&offset=${offset}`,
      "stmt_select_heavy",
      loop_stmt_select_heavy
    );
    recordSelect(
      `${BASE}/nonstmt/select/heavy?q=${encodeURIComponent(
        q
      )}&brand=${encodeURIComponent(
        brand
      )}&top_per_category=${topK}&limit=${limit}&offset=${offset}`,
      "nonstmt_select_heavy",
      loop_nonstmt_select_heavy
    );
  });

  group("insert_heavy", () => {
    // keep=false supaya tidak menumpuk data
    recordInsert(
      `${BASE}/stmt/insert/heavy?orders=${orders}&items_per_order=${itemsPerOrder}&q=${encodeURIComponent(
        q
      )}&brand=${encodeURIComponent(brand)}&keep=false&tx=true`,
      "stmt_insert_heavy",
      loop_stmt_insert_heavy
    );
    recordInsert(
      `${BASE}/nonstmt/insert/heavy?orders=${orders}&items_per_order=${itemsPerOrder}&q=${encodeURIComponent(
        q
      )}&brand=${encodeURIComponent(brand)}&keep=false&tx=false`,
      "nonstmt_insert_heavy",
      loop_nonstmt_insert_heavy
    );
  });

  const res = http.get(`${BASE}/health`);
  check(res, { "health is ok": (r) => r.status === 200 });

  sleep(0.2);
}

// ===== Helpers =====
function recordSelect(url, tagEndpoint, trendMetric) {
  const res = http.get(url, { tags: { endpoint: tagEndpoint } });
  check(res, { "select 200": (r) => r.status === 200 });

  const body = safeJSON(res);
  if (body && body.avg_per_loop) {
    const ms = toMs(body.avg_per_loop);
    if (ms !== null) trendMetric.add(ms);
  }
}

function recordInsert(url, tagEndpoint, trendMetric) {
  const res = http.get(url, { tags: { endpoint: tagEndpoint } });
  check(res, { "insert 200": (r) => r.status === 200 });

  const body = safeJSON(res);
  if (body && body.avg_per_loop) {
    const ms = toMs(body.avg_per_loop);
    if (ms !== null) trendMetric.add(ms);
  }
}

function safeJSON(res) {
  try {
    return res.json();
  } catch (_) {
    return null;
  }
}

// Convert duration → milliseconds (number)
function toMs(v) {
  if (typeof v === "number") return v / 1e6; // assume ns
  if (typeof v !== "string") return null;
  const s = v.trim().toLowerCase();
  if (s.endsWith("ns")) return Number(s.slice(0, -2)) / 1e6;
  if (s.endsWith("µs") || s.endsWith("us"))
    return Number(s.replace("µs", "us").slice(0, -2)) / 1e3;
  if (s.endsWith("ms")) return Number(s.slice(0, -2));
  if (s.endsWith("s")) return Number(s.slice(0, -1)) * 1000;
  const n = parseFloat(s);
  return Number.isFinite(n) ? n / 1e6 : null;
}

function pickOne(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}
function randInt(a, b) {
  return a + Math.floor(Math.random() * (b - a + 1));
}
