/**
 * Anthology Spor Motoru — 5 Büyük Platform Çözücüsü
 * 
 * 1. Taraftarium24 (teletv API)
 * 2. ArdaSports / SelcukSports (selcuksporkral1.cc + loadstream Base64)
 * 3. AlemdarSport (alemdarsport.cfd + mdalemdar CDN)
 * 4. MahsunSports (mahsunsports.xyz + Akinci IPTV CDN)
 * 5. PapazSports (papazsports.pro + auth.php + usertoken HLS)
 */

const { loadConfig, val } = require("./config.js");

let TARAFTARIUM_SITE = "https://taraftarium24-245.top/";
let TARAFTARIUM_API = "https://teletv15.top";
let TARAFTARIUM_BASE_NUM = 245;

let ARDA_SITE = "https://selcuksporkral1.cc";
let ALEMDAR_SITE = "https://alemdarsport.cfd";
let MAHSUN_SITE = "https://mahsunsports55.xyz";
let PAPAZ_SITE = "https://www.papazsports1028.pro";

const _HEADERS = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML like Gecko) Chrome/120.0.0.0 Safari/537.36',
    'Accept': '*/*'
};

let _cfgReady = null;
function initSportConfig() {
    if (!_cfgReady) {
        _cfgReady = loadConfig().then(function () {
            let v;
            v = val('urls.live.taraftarium.base'); if (v) TARAFTARIUM_SITE = String(v);
            v = val('urls.live.taraftarium.api_base'); if (v) TARAFTARIUM_API = String(v).replace(/\/+$/, '');
            v = val('urls.live.taraftarium.base_number'); if (v && !isNaN(Number(v))) TARAFTARIUM_BASE_NUM = Number(v);
            v = val('urls.live.ardasports.base'); if (v) ARDA_SITE = String(v).replace(/\/+$/, '');
            v = val('urls.live.alemdarsport.base'); if (v) ALEMDAR_SITE = String(v).replace(/\/+$/, '');
            v = val('urls.live.mahsunsports.base'); if (v) MAHSUN_SITE = String(v).replace(/\/+$/, '');
            v = val('urls.live.papazsports.base'); if (v) PAPAZ_SITE = String(v).replace(/\/+$/, '');
        });
    }
    return _cfgReady;
}

function cleanKey(s) {
    return (s || '').toString().toLowerCase().replace(/[^a-z0-9]/g, '');
}

function safeAtob(b64) {
    try {
        return Buffer.from(b64, 'base64').toString('utf8');
    } catch (e) {
        return '';
    }
}

function withTimeout(promise, ms) {
    let timer;
    const timeoutPromise = new Promise(function(resolve) {
        timer = setTimeout(function() { resolve(null); }, ms);
    });
    return Promise.race([
        promise.then(function(res) {
            clearTimeout(timer);
            return res;
        }).catch(function() {
            clearTimeout(timer);
            return null;
        }),
        timeoutPromise
    ]);
}

// ── Çoklu Platform Kanal Eşleştirme Sözlüğü ─────────────────────────
const SPORT_MAP = {
    "beinsports1": {
        name: "BEIN SPORTS 1 HD",
        taraftarium: "bein-sports-1",
        ardasports: "bein-sports-1",
        alemdarsport: "bein1",
        mahsunsports: "androstreamlivebs1",
        papazsports: "100001"
    },
    "beinsports2": {
        name: "BEIN SPORTS 2 HD",
        taraftarium: "bein-sports-2",
        ardasports: "bein-sports-2",
        alemdarsport: "bein2",
        mahsunsports: "androstreamlivebs2",
        papazsports: "100002"
    },
    "beinsports3": {
        name: "BEIN SPORTS 3 HD",
        taraftarium: "bein-sports-3",
        ardasports: "bein-sports-3",
        alemdarsport: "bein3",
        mahsunsports: "androstreamlivebs3",
        papazsports: "100003"
    },
    "beinsports4": {
        name: "BEIN SPORTS 4 HD",
        taraftarium: "bein-sports-4",
        ardasports: "bein-sports-4",
        alemdarsport: "bein4",
        mahsunsports: "androstreamlivebs4",
        papazsports: "100004"
    },
    "beinsports5": {
        name: "BEIN SPORTS 5 HD",
        taraftarium: "bein-sports-5",
        ardasports: "bein-sports-5",
        alemdarsport: "bein5",
        mahsunsports: "androstreamlivebs5",
        papazsports: "100005"
    },
    "beinsportsmax1": {
        name: "BEIN SPORTS MAX 1 HD",
        taraftarium: "bein-sports-max-1",
        ardasports: "bein-sports-max-1",
        alemdarsport: "beinmax1",
        mahsunsports: "androstreamlivebsm1",
        papazsports: "100006"
    },
    "beinsportsmax2": {
        name: "BEIN SPORTS MAX 2 HD",
        taraftarium: "bein-sports-max-2",
        ardasports: "bein-sports-max-2",
        alemdarsport: "beinmax2",
        mahsunsports: "androstreamlivebsm2",
        papazsports: "100007"
    },
    "ssport": {
        name: "S SPORT HD",
        taraftarium: "s-sport",
        ardasports: "s-sport",
        alemdarsport: "ssport1",
        mahsunsports: "androstreamlivess1",
        papazsports: "100010"
    },
    "ssport1": {
        name: "S SPORT 1 HD",
        taraftarium: "s-sport",
        ardasports: "s-sport",
        alemdarsport: "ssport1",
        mahsunsports: "androstreamlivess1",
        papazsports: "100010"
    },
    "ssport2": {
        name: "S SPORT 2 HD",
        taraftarium: "s-sport-2",
        ardasports: "s-sport-2",
        alemdarsport: "ssport2",
        mahsunsports: "androstreamlivess2",
        papazsports: "100011"
    },
    "ssportplus": {
        name: "S SPORT PLUS HD",
        mahsunsports: "androstreamlivessplus1",
        alemdarsport: "ssport1"
    },
    "tivibuspor": {
        name: "TIVIBU SPOR HD",
        alemdarsport: "tivibu1",
        mahsunsports: "androstreamlivets",
        papazsports: "100021"
    },
    "tivibuspor1": {
        name: "TIVIBU SPOR 1 HD",
        alemdarsport: "tivibu1",
        mahsunsports: "androstreamlivets1",
        papazsports: "100021"
    },
    "tivibuspor2": {
        name: "TIVIBU SPOR 2 HD",
        alemdarsport: "tivibu2",
        mahsunsports: "androstreamlivets2",
        papazsports: "100022"
    },
    "tivibuspor3": {
        name: "TIVIBU SPOR 3 HD",
        alemdarsport: "tivibu3",
        mahsunsports: "androstreamlivets3",
        papazsports: "100023"
    },
    "tivibuspor4": {
        name: "TIVIBU SPOR 4 HD",
        alemdarsport: "tivibu4",
        mahsunsports: "androstreamlivets4",
        papazsports: "100024"
    },
    "smartspor1": {
        name: "SMART SPOR 1 HD",
        alemdarsport: "sporsmart",
        mahsunsports: "androstreamlivesm1",
        papazsports: "100030"
    },
    "smartspor2": {
        name: "SMART SPOR 2 HD",
        alemdarsport: "sporsmart2",
        mahsunsports: "androstreamlivesm2",
        papazsports: "100031"
    },
    "trtspor": {
        name: "TRT SPOR HD",
        taraftarium: "trt-spor",
        ardasports: "trt-spor",
        alemdarsport: "trtspor"
    },
    "trtsporyildiz": {
        name: "TRT SPOR YILDIZ HD",
        alemdarsport: "trtspor2"
    },
    "trt1": {
        name: "TRT 1 HD",
        taraftarium: "trt-1",
        ardasports: "trt-1",
        alemdarsport: "trt1"
    },
    "aspor": {
        name: "A SPOR HD",
        taraftarium: "a-spor",
        ardasports: "a-spor",
        alemdarsport: "aspor",
        papazsports: "100052"
    },
    "tv85": {
        name: "TV8.5 HD",
        alemdarsport: "tv85",
        papazsports: "100051"
    },
    "nbatv": {
        name: "NBA TV HD",
        papazsports: "100053"
    },
    "eurosport1": {
        name: "EUROSPORT 1 HD",
        mahsunsports: "androstreamlivees1",
        papazsports: "100056"
    },
    "eurosport2": {
        name: "EUROSPORT 2 HD",
        mahsunsports: "androstreamlivees2",
        papazsports: "100057"
    },
    "cbcsport": {
        name: "CBC SPORT HD",
        mahsunsports: "androstreamlivecbcs"
    },
    "idmantv": {
        name: "IDMAN TV HD",
        mahsunsports: "androstreamliveidm"
    },
    "exxentv": {
        name: "EXXEN TV HD",
        mahsunsports: "androstreamliveexn"
    }
};

function findSportChannelMapping(rawStr) {
    if (!rawStr) return null;
    const k = cleanKey(rawStr).replace(/^tv/, '').replace(/tr$/, '');
    if (SPORT_MAP[k]) return SPORT_MAP[k];

    if (k.includes('beinsportsmax1') || k.includes('beinmax1')) return SPORT_MAP['beinsportsmax1'];
    if (k.includes('beinsportsmax2') || k.includes('beinmax2')) return SPORT_MAP['beinsportsmax2'];
    if (k.includes('beinsport1') || k.includes('bein1')) return SPORT_MAP['beinsports1'];
    if (k.includes('beinsport2') || k.includes('bein2')) return SPORT_MAP['beinsports2'];
    if (k.includes('beinsport3') || k.includes('bein3')) return SPORT_MAP['beinsports3'];
    if (k.includes('beinsport4') || k.includes('bein4')) return SPORT_MAP['beinsports4'];
    if (k.includes('beinsport5') || k.includes('bein5')) return SPORT_MAP['beinsports5'];
    if (k.includes('ssportplus')) return SPORT_MAP['ssportplus'];
    if (k.includes('ssport2')) return SPORT_MAP['ssport2'];
    if (k.includes('ssport')) return SPORT_MAP['ssport1'];
    if (k.includes('tivibu1')) return SPORT_MAP['tivibuspor1'];
    if (k.includes('tivibu2')) return SPORT_MAP['tivibuspor2'];
    if (k.includes('tivibu3')) return SPORT_MAP['tivibuspor3'];
    if (k.includes('tivibu4')) return SPORT_MAP['tivibuspor4'];
    if (k.includes('tivibu')) return SPORT_MAP['tivibuspor'];
    if (k.includes('smartspor2') || k.includes('sporsmart2')) return SPORT_MAP['smartspor2'];
    if (k.includes('smartspor') || k.includes('sporsmart')) return SPORT_MAP['smartspor1'];
    if (k.includes('trtsporyildiz')) return SPORT_MAP['trtsporyildiz'];
    if (k.includes('trtspor')) return SPORT_MAP['trtspor'];
    if (k.includes('aspor')) return SPORT_MAP['aspor'];
    if (k.includes('tv85') || k.includes('tv8bucuk')) return SPORT_MAP['tv85'];
    if (k.includes('nba')) return SPORT_MAP['nbatv'];
    if (k.includes('eurosport2')) return SPORT_MAP['eurosport2'];
    if (k.includes('eurosport')) return SPORT_MAP['eurosport1'];
    if (k.includes('cbc')) return SPORT_MAP['cbcsport'];
    if (k.includes('idman')) return SPORT_MAP['idmantv'];
    if (k.includes('exxen')) return SPORT_MAP['exxentv'];

    for (const mk in SPORT_MAP) {
        if (k.startsWith(mk) || mk.startsWith(k)) return SPORT_MAP[mk];
    }
    return null;
}

// ── 1. Taraftarium24 Çözücü ──────────────────────────────────────────
let _resolvedT24Api = null;
let _resolvedT24Site = null;
let _resolvedT24Time = 0;
const RESOLVER_TTL = 15 * 60 * 1000;

async function resolveTaraftariumBase() {
    const now = Date.now();
    if (_resolvedT24Api && (now - _resolvedT24Time < RESOLVER_TTL)) {
        return { api: _resolvedT24Api, site: _resolvedT24Site };
    }

    const baseApi = _resolvedT24Api || TARAFTARIUM_API;
    const baseSite = _resolvedT24Site || TARAFTARIUM_SITE;

    try {
        const res = await fetch(baseApi + "/load/channels.php", {
            headers: _HEADERS,
            signal: AbortSignal.timeout(3000)
        });
        if (res && res.status === 200) {
            _resolvedT24Api = baseApi;
            _resolvedT24Site = baseSite;
            _resolvedT24Time = now;
            return { api: _resolvedT24Api, site: _resolvedT24Site };
        }
    } catch (e) {}

    const num = TARAFTARIUM_BASE_NUM;
    const cands = [num, num + 1, num - 1, num + 2, num + 3, num + 4];

    for (const curN of cands) {
        const candUrl = `https://taraftarium24-${curN}.top/`;
        try {
            const r = await fetch(candUrl, { headers: _HEADERS, signal: AbortSignal.timeout(2500) });
            const html = await r.text();
            if (!html) continue;
            const apiM = html.match(/https:\/\/teletv\d+\.top/i) || html.match(/fetch\(['"](https:\/\/[^'"]*teletv[^'"]*top)/i);
            const foundApi = apiM ? (apiM[1] || apiM[0]) : baseApi;
            _resolvedT24Api = foundApi.replace(/\/+$/, '');
            _resolvedT24Site = candUrl;
            _resolvedT24Time = now;
            return { api: _resolvedT24Api, site: _resolvedT24Site };
        } catch (e) {}
    }

    _resolvedT24Api = baseApi;
    _resolvedT24Site = baseSite;
    _resolvedT24Time = now;
    return { api: _resolvedT24Api, site: _resolvedT24Site };
}

async function fetchTaraftariumStream(channelId) {
    if (!channelId) return null;
    try {
        const base = await resolveTaraftariumBase();
        const apiUrl = base.api + "/load/yayinlink.php?id=" + encodeURIComponent(channelId);
        const res = await fetch(apiUrl, { headers: _HEADERS, signal: AbortSignal.timeout(3500) });
        const data = await res.json();
        if (data && data.deismackanal && typeof data.deismackanal === 'string' && data.deismackanal.includes('m3u8')) {
            const finalUrl = data.deismackanal.replace(/edge4\./g, 'edge3.');
            return {
                url: finalUrl,
                headers: {
                    'User-Agent': _HEADERS['User-Agent'],
                    'Referer': base.site,
                    'Origin': base.site
                }
            };
        }
    } catch (e) {}
    return null;
}

// ── 2. ArdaSports (SelcukSports) Çözücü ──────────────────────────────
async function fetchArdaSportsStream(channelId) {
    if (!channelId) return null;
    try {
        const watchUrl = ARDA_SITE + "/channel/watch/" + encodeURIComponent(channelId);
        const refHeader = ARDA_SITE + "/canli-izle/" + encodeURIComponent(channelId);

        const res = await fetch(watchUrl, {
            headers: {
                'User-Agent': _HEADERS['User-Agent'],
                'Referer': refHeader
            },
            signal: AbortSignal.timeout(3500)
        });
        const html = await res.text();
        const lsMatch = html.match(/src="(\/loadstream\.php\?[^"]+)"/);
        if (!lsMatch) return null;

        const loadUrl = ARDA_SITE + lsMatch[1];
        const lsRes = await fetch(loadUrl, {
            headers: {
                'User-Agent': _HEADERS['User-Agent'],
                'Referer': watchUrl
            },
            signal: AbortSignal.timeout(3500)
        });
        const lsHtml = await lsRes.text();
        const b64Match = lsHtml.match(/['"]([A-Za-z0-9+/=]{40,})\|atob/i) || lsHtml.match(/atob\(['"]([A-Za-z0-9+/=]+)['"]\)/);
        if (!b64Match) return null;

        const decoded = safeAtob(b64Match[1]);
        const streamM = decoded.match(/https?:\/\/[^"'\s\\]+\.m3u8/);
        if (streamM) {
            return {
                url: streamM[0],
                headers: {
                    'User-Agent': _HEADERS['User-Agent'],
                    'Referer': ARDA_SITE,
                    'Origin': ARDA_SITE
                }
            };
        }
    } catch (e) {}
    return null;
}

// ── 3. AlemdarSport Çözücü ──────────────────────────────────────────
let _cachedAlemdarCdn = null;
let _cachedAlemdarTime = 0;

async function fetchAlemdarStream(channelId) {
    if (!channelId) return null;
    const now = Date.now();
    if (_cachedAlemdarCdn && (now - _cachedAlemdarTime < RESOLVER_TTL)) {
        return {
            url: _cachedAlemdarCdn.replace('${cid}', channelId),
            headers: {
                'User-Agent': _HEADERS['User-Agent'],
                'Referer': ALEMDAR_SITE,
                'Origin': ALEMDAR_SITE
            }
        };
    }

    try {
        const channelUrl = ALEMDAR_SITE + "/channel.php?id=" + encodeURIComponent(channelId);
        const res = await fetch(channelUrl, { headers: _HEADERS, signal: AbortSignal.timeout(3500) });
        const html = await res.text();
        const flotfMatch = html.match(/flotfUrl\s*=\s*[`'"](https:\/\/[^`'"]+index\.m3u8)/);
        if (flotfMatch) {
            const template = flotfMatch[1];
            _cachedAlemdarCdn = template;
            _cachedAlemdarTime = now;
            return {
                url: template.replace('${cid}', channelId),
                headers: {
                    'User-Agent': _HEADERS['User-Agent'],
                    'Referer': ALEMDAR_SITE,
                    'Origin': ALEMDAR_SITE
                }
            };
        }
    } catch (e) {}
    return null;
}

// ── 4. MahsunSports Çözücü ──────────────────────────────────────────
let _cachedMahsunBase = null;
let _cachedMahsunTime = 0;

async function resolveMahsunBaseUrl() {
    const now = Date.now();
    if (_cachedMahsunBase && (now - _cachedMahsunTime < RESOLVER_TTL)) {
        return _cachedMahsunBase;
    }

    const probeUrl = MAHSUN_SITE + "/event.html?id=androstreamlivebs1";
    try {
        const res = await fetch(probeUrl, { headers: _HEADERS, signal: AbortSignal.timeout(3500) });
        const html = await res.text();
        const baseMatch = html.match(/["'](https:\/\/andro\.[^"']+\/checklist\/)["']/);
        if (baseMatch) {
            _cachedMahsunBase = baseMatch[1];
            _cachedMahsunTime = now;
            return _cachedMahsunBase;
        }
    } catch (e) {}

    const fallback = "https://andro.evrenesoglu157.click/checklist/";
    _cachedMahsunBase = fallback;
    _cachedMahsunTime = now;
    return fallback;
}

async function fetchMahsunStream(channelId) {
    if (!channelId) return null;
    try {
        const base = await resolveMahsunBaseUrl();
        return {
            url: base + channelId + ".m3u8",
            headers: {
                'User-Agent': _HEADERS['User-Agent']
            }
        };
    } catch (e) {}
    return null;
}

// ── 5. PapazSports Çözücü ───────────────────────────────────────────
async function fetchPapazStream(channelId) {
    if (!channelId) return null;
    try {
        const authUrl = PAPAZ_SITE + "/auth.php";
        const res = await fetch(authUrl, {
            method: 'POST',
            headers: {
                'User-Agent': _HEADERS['User-Agent'],
                'Referer': PAPAZ_SITE + '/',
                'Origin': PAPAZ_SITE,
                'X-Requested-With': 'XMLHttpRequest',
                'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8'
            },
            body: 'channel=' + encodeURIComponent(channelId),
            signal: AbortSignal.timeout(3500)
        });
        const json = await res.json();
        if (json && json.URL && json.TOKEN) {
            return {
                url: json.URL,
                headers: {
                    'User-Agent': _HEADERS['User-Agent'],
                    'Referer': PAPAZ_SITE + '/',
                    'Origin': PAPAZ_SITE,
                    'usertoken': json.TOKEN
                }
            };
        }
    } catch (e) {}
    return null;
}

/**
 * 5 Büyük Platform Eşzamanlı Çözümleme
 */
async function resolveBestSportStream(channelIdOrName) {
    await initSportConfig();
    const map = findSportChannelMapping(channelIdOrName);
    if (!map) return null;

    const pTimeout = 3800;
    const tPromise = map.taraftarium ? withTimeout(fetchTaraftariumStream(map.taraftarium), pTimeout) : Promise.resolve(null);
    const aPromise = map.ardasports ? withTimeout(fetchArdaSportsStream(map.ardasports), pTimeout) : Promise.resolve(null);
    const mPromise = map.mahsunsports ? withTimeout(fetchMahsunStream(map.mahsunsports), pTimeout) : Promise.resolve(null);
    const alPromise = map.alemdarsport ? withTimeout(fetchAlemdarStream(map.alemdarsport), pTimeout) : Promise.resolve(null);
    const pPromise = map.papazsports ? withTimeout(fetchPapazStream(map.papazsports), pTimeout) : Promise.resolve(null);

    const results = await Promise.all([aPromise, tPromise, mPromise, alPromise, pPromise]);

    // ArdaSports > Taraftarium > MahsunSports > Alemdar > Papaz
    for (const r of results) {
        if (r && r.url) {
            return {
                channelName: map.name,
                url: r.url,
                headers: r.headers || {}
            };
        }
    }
    return null;
}

module.exports = {
    SPORT_MAP,
    findSportChannelMapping,
    resolveBestSportStream
};
