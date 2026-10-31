const http = require('http');
const os = require('os');
const { SPORT_MAP, findSportChannelMapping, resolveBestSportStream } = require('./sport_engine.js');

const PORT = process.env.PORT || 3555;

function getLocalIp() {
    const interfaces = os.networkInterfaces();
    for (const name of Object.keys(interfaces)) {
        for (const iface of interfaces[name]) {
            if (iface.family === 'IPv4' && !iface.internal) {
                return iface.address;
            }
        }
    }
    return '127.0.0.1';
}

function b64Encode(str) {
    return Buffer.from(str || '', 'utf8').toString('base64url');
}

function b64Decode(str) {
    return Buffer.from(str || '', 'base64url').toString('utf8');
}

// In-memory stream resolver cache (60s TTL)
const _streamCache = new Map();

async function getCachedStream(channelKey) {
    const now = Date.now();
    const cached = _streamCache.get(channelKey);
    if (cached && (now - cached.time < 60000)) {
        return cached.stream;
    }

    const stream = await resolveBestSportStream(channelKey);
    if (stream && stream.url) {
        _streamCache.set(channelKey, { stream, time: now });
    }
    return stream;
}

// In-memory segment cache (15s TTL)
const _segmentCache = new Map();

function cleanSegmentCache() {
    const now = Date.now();
    for (const [k, v] of _segmentCache.entries()) {
        if (now - v.time > 20000) {
            _segmentCache.delete(k);
        }
    }
}
setInterval(cleanSegmentCache, 10000);

// Generate M3U Playlist
function generateM3U(baseUrl) {
    let m3u = '#EXTM3U x-tvg-url="https://epg.share247.net/epg.xml.gz"\n\n';

    const logos = {
        beinsports1: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/beinsports.png",
        beinsports2: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/beinsports.png",
        beinsports3: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/beinsports.png",
        beinsports4: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/beinsports.png",
        beinsports5: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/beinsports.png",
        beinsportsmax1: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/beinsports.png",
        beinsportsmax2: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/beinsports.png",
        ssport1: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/ssport.png",
        ssport2: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/ssport.png",
        ssportplus: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/ssport.png",
        tivibuspor1: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/tivibuspor.png",
        tivibuspor2: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/tivibuspor.png",
        tivibuspor3: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/tivibuspor.png",
        tivibuspor4: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/tivibuspor.png",
        smartspor1: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/smartspor.png",
        smartspor2: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/smartspor.png",
        trtspor: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/trtspor.png",
        trtsporyildiz: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/trtsporyildiz.png",
        trt1: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/trt1.png",
        aspor: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/aspor.png",
        tv85: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/tv85.png",
        nbatv: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/nbatv.png",
        eurosport1: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/eurosport.png",
        eurosport2: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/eurosport.png",
        cbcsport: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/cbcsport.png",
        idmantv: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/idmantv.png",
        exxentv: "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/exxen.png"
    };

    const orderedKeys = [
        "beinsports1", "beinsports2", "beinsports3", "beinsports4", "beinsports5",
        "beinsportsmax1", "beinsportsmax2",
        "ssport1", "ssport2", "ssportplus",
        "tivibuspor1", "tivibuspor2", "tivibuspor3", "tivibuspor4",
        "smartspor1", "smartspor2",
        "trtspor", "trtsporyildiz", "trt1", "aspor", "tv85",
        "nbatv", "eurosport1", "eurosport2", "cbcsport", "idmantv", "exxentv"
    ];

    for (const key of orderedKeys) {
        const item = SPORT_MAP[key];
        if (!item) continue;
        const logo = logos[key] || "https://raw.githubusercontent.com/falsisdev/anthology/main/assets/canli/default_tv.png";
        m3u += `#EXTINF:-1 tvg-id="${key}.tr" tvg-name="${item.name}" tvg-logo="${logo}" group-title="Spor Kanalları",${item.name}\n`;
        m3u += `${baseUrl}/live/${key}.m3u8\n\n`;
    }

    return m3u;
}

const server = http.createServer(async (req, res) => {
    // CORS headers
    res.setHeader('Access-Control-Allow-Origin', '*');
    res.setHeader('Access-Control-Allow-Methods', 'GET, HEAD, OPTIONS');
    res.setHeader('Access-Control-Allow-Headers', '*');

    if (req.method === 'OPTIONS') {
        res.writeHead(200);
        res.end();
        return;
    }

    const hostHeader = req.headers.host || `127.0.0.1:${PORT}`;
    const baseUrl = `http://${hostHeader}`;
    const parsedUrl = new URL(req.url, baseUrl);
    const pathname = parsedUrl.pathname;

    // ── 1. Web Dashboard (Landing Page) ──
    if (pathname === '/' || pathname === '/index.html') {
        const localIp = getLocalIp();
        const iptvUrl = `http://${localIp}:${PORT}/iptv.m3u`;
        const localIptvUrl = `http://127.0.0.1:${PORT}/iptv.m3u`;

        res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
        res.end(`<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Anthology Live — Kesintisiz Spor & IPTV</title>
    <style>
        :root {
            --bg: #0b0f19;
            --card-bg: rgba(255, 255, 255, 0.04);
            --card-border: rgba(255, 255, 255, 0.08);
            --accent: #3b82f6;
            --accent-glow: rgba(59, 130, 246, 0.35);
            --text-main: #f8fafc;
            --text-sub: #94a3b8;
            --green: #10b981;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, "SF Pro Display", "SF Pro Text", "Segoe UI", Roboto, sans-serif; }
        body { background: var(--bg); color: var(--text-main); min-height: 100vh; padding: 40px 20px; display: flex; flex-direction: column; align-items: center; }
        .container { max-width: 840px; width: 100%; }
        header { text-align: center; margin-bottom: 35px; }
        .badge { display: inline-flex; align-items: center; gap: 6px; padding: 4px 12px; border-radius: 999px; background: rgba(16, 185, 129, 0.15); color: var(--green); font-size: 13px; font-weight: 600; margin-bottom: 12px; border: 1px solid rgba(16, 185, 129, 0.3); }
        .badge::before { content: ""; width: 8px; height: 8px; border-radius: 50%; background: var(--green); box-shadow: 0 0 8px var(--green); }
        h1 { font-size: 32px; font-weight: 700; letter-spacing: -0.5px; margin-bottom: 8px; }
        p.subtitle { color: var(--text-sub); font-size: 15px; }
        .box { background: var(--card-bg); border: 1px solid var(--card-border); border-radius: 16px; padding: 24px; margin-bottom: 24px; backdrop-filter: blur(20px); -webkit-backdrop-filter: blur(20px); }
        .box-title { font-size: 16px; font-weight: 600; margin-bottom: 12px; display: flex; align-items: center; gap: 8px; }
        .url-row { display: flex; gap: 10px; background: rgba(0, 0, 0, 0.35); border: 1px solid rgba(255, 255, 255, 0.05); padding: 10px 14px; border-radius: 10px; align-items: center; justify-content: space-between; word-break: break-all; font-family: monospace; font-size: 14px; color: #60a5fa; }
        button.copy-btn { background: var(--accent); color: #fff; border: none; padding: 8px 16px; border-radius: 8px; font-size: 13px; font-weight: 600; cursor: pointer; transition: all 0.2s; white-space: nowrap; }
        button.copy-btn:hover { background: #2563eb; transform: translateY(-1px); }
        .channels-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 12px; margin-top: 15px; }
        .channel-card { background: rgba(255, 255, 255, 0.02); border: 1px solid rgba(255, 255, 255, 0.06); padding: 14px; border-radius: 12px; display: flex; align-items: center; justify-content: space-between; text-decoration: none; color: inherit; transition: all 0.2s; }
        .channel-card:hover { background: rgba(255, 255, 255, 0.06); border-color: rgba(59, 130, 246, 0.4); transform: translateY(-2px); }
        .ch-name { font-size: 14px; font-weight: 600; }
        .ch-tag { font-size: 11px; color: #10b981; background: rgba(16, 185, 129, 0.1); padding: 2px 8px; border-radius: 6px; font-weight: 500; }
        .guide-step { display: flex; gap: 12px; margin-bottom: 14px; font-size: 14px; line-height: 1.5; color: #cbd5e1; }
        .step-num { width: 24px; height: 24px; border-radius: 50%; background: var(--accent); color: white; display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700; flex-shrink: 0; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <div class="badge">CANLI YAYIN PROXY AKTİF</div>
            <h1>Anthology Live</h1>
            <p class="subtitle">SenPlayer, Infuse ve Apple TV için Sıfır Kasma & Tek Parça Canlı Akış Servisi</p>
        </header>

        <div class="box">
            <div class="box-title">📺 SenPlayer / IPTV Çalma Listesi URL (M3U)</div>
            <p style="font-size:13px; color:var(--text-sub); margin-bottom: 12px;">SenPlayer uygulamasında <strong>IPTV &gt; + &gt; M3U URL Ekle</strong> kısmına bu adresi yapıştırın:</p>
            <div class="url-row">
                <span id="iptvUrl">${iptvUrl}</span>
                <button class="copy-btn" onclick="copyUrl('${iptvUrl}', this)">Kopyala</button>
            </div>
            <div style="margin-top: 10px; font-size: 12px; color: #64748b;">
                Yerel (Mac üzerinde SenPlayer için): <a href="/iptv.m3u" style="color: #94a3b8;">${localIptvUrl}</a>
            </div>
        </div>

        <div class="box">
            <div class="box-title">⚡️ SenPlayer Donma ve Parçalanma Önleyici Ayarı</div>
            <div class="guide-step">
                <div class="step-num">1</div>
                <div>SenPlayer uygulamasını açın ve <strong>Ayarlar &gt; Oynatma / Oynatıcı Çekirdeği</strong> seçeneğine gidin.</div>
            </div>
            <div class="guide-step">
                <div class="step-num">2</div>
                <div>Oynatıcı Çekirdeğini <strong>KSPlayer (veya FFmpeg)</strong> olarak seçin. Bu sayede tüm TS parçaları tek bir kesintisiz livestream olarak oynatılır.</div>
            </div>
            <div class="guide-step">
                <div class="step-num">3</div>
                <div>IPTV sekmesinden <strong>Anthology Live</strong> listesini açıp dilediğiniz beIN Sports veya spor kanalını başlatın.</div>
            </div>
        </div>

        <div class="box">
            <div class="box-title">⚽️ Hazır Canlı Spor Kanalları</div>
            <div class="channels-grid">
                <a class="channel-card" href="/live/beinsports1.m3u8" target="_blank"><span class="ch-name">BeIN Sports 1 HD</span><span class="ch-tag">1080p</span></a>
                <a class="channel-card" href="/live/beinsports2.m3u8" target="_blank"><span class="ch-name">BeIN Sports 2 HD</span><span class="ch-tag">1080p</span></a>
                <a class="channel-card" href="/live/beinsports3.m3u8" target="_blank"><span class="ch-name">BeIN Sports 3 HD</span><span class="ch-tag">1080p</span></a>
                <a class="channel-card" href="/live/beinsports4.m3u8" target="_blank"><span class="ch-name">BeIN Sports 4 HD</span><span class="ch-tag">1080p</span></a>
                <a class="channel-card" href="/live/beinsports5.m3u8" target="_blank"><span class="ch-name">BeIN Sports 5 HD</span><span class="ch-tag">1080p</span></a>
                <a class="channel-card" href="/live/beinsportsmax1.m3u8" target="_blank"><span class="ch-name">BeIN Sports Max 1</span><span class="ch-tag">1080p</span></a>
                <a class="channel-card" href="/live/beinsportsmax2.m3u8" target="_blank"><span class="ch-name">BeIN Sports Max 2</span><span class="ch-tag">1080p</span></a>
                <a class="channel-card" href="/live/ssport1.m3u8" target="_blank"><span class="ch-name">S Sport 1 HD</span><span class="ch-tag">1080p</span></a>
                <a class="channel-card" href="/live/ssport2.m3u8" target="_blank"><span class="ch-name">S Sport 2 HD</span><span class="ch-tag">1080p</span></a>
                <a class="channel-card" href="/live/tivibuspor1.m3u8" target="_blank"><span class="ch-name">Tivibu Spor 1</span><span class="ch-tag">1080p</span></a>
                <a class="channel-card" href="/live/smartspor1.m3u8" target="_blank"><span class="ch-name">Smart Spor 1 HD</span><span class="ch-tag">1080p</span></a>
                <a class="channel-card" href="/live/trtspor.m3u8" target="_blank"><span class="ch-name">TRT Spor HD</span><span class="ch-tag">1080p</span></a>
            </div>
        </div>
    </div>

    <script>
        function copyUrl(text, btn) {
            navigator.clipboard.writeText(text).then(() => {
                const orig = btn.innerText;
                btn.innerText = 'Kopyalandı!';
                btn.style.background = '#10b981';
                setTimeout(() => {
                    btn.innerText = orig;
                    btn.style.background = '#3b82f6';
                }, 2000);
            });
        }
    </script>
</body>
</html>`);
        return;
    }

    // ── 2. IPTV Playlist (M3U) ──
    if (pathname === '/iptv.m3u' || pathname === '/playlist.m3u' || pathname === '/canli.m3u') {
        const m3uContent = generateM3U(baseUrl);
        res.writeHead(200, {
            'Content-Type': 'application/vnd.apple.mpegurl; charset=utf-8',
            'Content-Disposition': 'inline; filename="anthology_live.m3u"',
            'Cache-Control': 'no-cache, no-store, must-revalidate'
        });
        res.end(m3uContent);
        return;
    }

    // ── 3. Live Stream HLS Playlist Proxy (/live/:channelId.m3u8) ──
    const liveMatch = pathname.match(/^\/live\/([^.]+)\.m3u8$/);
    if (liveMatch) {
        const channelKey = liveMatch[1].toLowerCase();
        try {
            const stream = await getCachedStream(channelKey);
            if (!stream || !stream.url) {
                res.writeHead(404, { 'Content-Type': 'text/plain; charset=utf-8' });
                res.end(`Kanal akışı bulunamadı: ${channelKey}`);
                return;
            }

            const upstreamHeaders = {
                'User-Agent': stream.headers['User-Agent'] || 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
                'Accept': '*/*'
            };
            if (stream.headers['Referer']) upstreamHeaders['Referer'] = stream.headers['Referer'];
            if (stream.headers['Origin']) upstreamHeaders['Origin'] = stream.headers['Origin'];
            if (stream.headers['usertoken']) upstreamHeaders['usertoken'] = stream.headers['usertoken'];

            const upRes = await fetch(stream.url, {
                headers: upstreamHeaders,
                signal: AbortSignal.timeout(5000)
            });

            if (!upRes.ok) {
                _streamCache.delete(channelKey); // Invalidate dead cache
                res.writeHead(502, { 'Content-Type': 'text/plain; charset=utf-8' });
                res.end(`Upstream CDN hata döndürdü: ${upRes.status}`);
                return;
            }

            const manifestText = await upRes.text();
            const upstreamBaseUrl = new URL('.', stream.url).href;

            // Rewrite playlist lines
            const lines = manifestText.split('\n');
            const rewritten = [];

            for (let i = 0; i < lines.length; i++) {
                const line = lines[i].trim();
                if (!line) {
                    rewritten.push(lines[i]);
                    continue;
                }

                if (line.startsWith('#')) {
                    // Check for sub-manifests inside EXT-X-STREAM-INF
                    if (line.startsWith('#EXT-X-STREAM-INF')) {
                        rewritten.push(line);
                        i++;
                        if (i < lines.length) {
                            const subLine = lines[i].trim();
                            const subAbsolute = new URL(subLine, upstreamBaseUrl).href;
                            const proxiedSub = `${baseUrl}/sub_manifest?u=${b64Encode(subAbsolute)}&r=${b64Encode(stream.headers['Referer'] || '')}&ua=${b64Encode(stream.headers['User-Agent'] || '')}`;
                            rewritten.push(proxiedSub);
                        }
                        continue;
                    }
                    rewritten.push(line);
                } else {
                    // This is a media segment URL (.ts, .jpg, .m4s etc.)
                    const segmentAbsolute = new URL(line, upstreamBaseUrl).href;
                    const proxiedSegment = `${baseUrl}/segment?u=${b64Encode(segmentAbsolute)}&r=${b64Encode(stream.headers['Referer'] || '')}&ua=${b64Encode(stream.headers['User-Agent'] || '')}`;
                    rewritten.push(proxiedSegment);
                }
            }

            const outputManifest = rewritten.join('\n');
            res.writeHead(200, {
                'Content-Type': 'application/vnd.apple.mpegurl; charset=utf-8',
                'Cache-Control': 'no-cache, no-store, must-revalidate',
                'Access-Control-Allow-Origin': '*'
            });
            res.end(outputManifest);
            return;
        } catch (err) {
            _streamCache.delete(channelKey);
            res.writeHead(500, { 'Content-Type': 'text/plain; charset=utf-8' });
            res.end(`Kanal çözme hatası: ${err.message}`);
            return;
        }
    }

    // ── 4. Sub-manifest Proxy (/sub_manifest) ──
    if (pathname === '/sub_manifest') {
        const targetUrl = b64Decode(parsedUrl.searchParams.get('u'));
        const referer = b64Decode(parsedUrl.searchParams.get('r'));
        const ua = b64Decode(parsedUrl.searchParams.get('ua')) || 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36';

        if (!targetUrl) {
            res.writeHead(400, { 'Content-Type': 'text/plain' });
            res.end('Missing URL');
            return;
        }

        try {
            const upHeaders = { 'User-Agent': ua, 'Accept': '*/*' };
            if (referer) {
                upHeaders['Referer'] = referer;
                upHeaders['Origin'] = referer;
            }

            const upRes = await fetch(targetUrl, { headers: upHeaders, signal: AbortSignal.timeout(5000) });
            const manifestText = await upRes.text();
            const upstreamBaseUrl = new URL('.', targetUrl).href;

            const lines = manifestText.split('\n');
            const rewritten = [];

            for (const line of lines) {
                const trimmed = line.trim();
                if (!trimmed || trimmed.startsWith('#')) {
                    rewritten.push(line);
                } else {
                    const segmentAbsolute = new URL(trimmed, upstreamBaseUrl).href;
                    const proxiedSegment = `${baseUrl}/segment?u=${b64Encode(segmentAbsolute)}&r=${b64Encode(referer)}&ua=${b64Encode(ua)}`;
                    rewritten.push(proxiedSegment);
                }
            }

            res.writeHead(200, {
                'Content-Type': 'application/vnd.apple.mpegurl; charset=utf-8',
                'Cache-Control': 'no-cache, no-store, must-revalidate',
                'Access-Control-Allow-Origin': '*'
            });
            res.end(rewritten.join('\n'));
            return;
        } catch (err) {
            res.writeHead(502, { 'Content-Type': 'text/plain' });
            res.end(`Submanifest error: ${err.message}`);
            return;
        }
    }

    // ── 5. Segment Proxy (/segment) ──
    if (pathname === '/segment') {
        const targetUrl = b64Decode(parsedUrl.searchParams.get('u'));
        const referer = b64Decode(parsedUrl.searchParams.get('r'));
        const ua = b64Decode(parsedUrl.searchParams.get('ua')) || 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36';

        if (!targetUrl) {
            res.writeHead(400, { 'Content-Type': 'text/plain' });
            res.end('Missing segment URL');
            return;
        }

        // Check memory cache
        const cachedSeg = _segmentCache.get(targetUrl);
        if (cachedSeg) {
            res.writeHead(200, {
                'Content-Type': 'video/mp2t',
                'Content-Length': cachedSeg.buffer.length,
                'Accept-Ranges': 'bytes',
                'Access-Control-Allow-Origin': '*',
                'Cache-Control': 'public, max-age=10'
            });
            res.end(cachedSeg.buffer);
            return;
        }

        try {
            const upHeaders = { 'User-Agent': ua, 'Accept': '*/*' };
            if (referer) {
                upHeaders['Referer'] = referer;
                upHeaders['Origin'] = referer;
            }

            const upRes = await fetch(targetUrl, { headers: upHeaders, signal: AbortSignal.timeout(15000) });
            if (!upRes.ok) {
                res.writeHead(upRes.status, { 'Content-Type': 'text/plain' });
                res.end(`Segment fetch failed: ${upRes.status}`);
                return;
            }

            const contentLength = upRes.headers.get('content-length');
            const resHeaders = {
                'Content-Type': 'video/mp2t',
                'Accept-Ranges': 'bytes',
                'Access-Control-Allow-Origin': '*',
                'Cache-Control': 'public, max-age=30'
            };
            if (contentLength) resHeaders['Content-Length'] = contentLength;

            res.writeHead(200, resHeaders);
            const { Readable } = require('stream');
            Readable.fromWeb(upRes.body).pipe(res);
            return;
        } catch (err) {
            res.writeHead(502, { 'Content-Type': 'text/plain' });
            res.end(`Segment error: ${err.message}`);
            return;
        }
    }

    // 404
    res.writeHead(404, { 'Content-Type': 'text/plain; charset=utf-8' });
    res.end('Sayfa bulunamadı.');
});

server.listen(PORT, '0.0.0.0', () => {
    const ip = getLocalIp();
    console.log('============================================================');
    console.log('🚀 Anthology Live Proxy Aktif!');
    console.log('============================================================');
    console.log(`📱 SenPlayer / IPTV M3U URL: http://${ip}:${PORT}/iptv.m3u`);
    console.log(`💻 Yerel URL (Tarayıcı / Mac): http://localhost:${PORT}/`);
    console.log('============================================================');
});

module.exports = server;
