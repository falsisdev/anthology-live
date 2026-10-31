const fs = require('fs');
const path = require('path');

const REMOTE_CONFIG_URL = "https://raw.githubusercontent.com/falsisdev/anthology/main/config.json";
const TTL_MS = 10 * 60 * 1000;

let _config = null;
let _lastLoad = 0;

function readLocalConfig() {
    try {
        const p = path.join(__dirname, 'config.json');
        if (fs.existsSync(p)) {
            return JSON.parse(fs.readFileSync(p, 'utf8'));
        }
    } catch (e) {}
    return null;
}

async function loadConfig() {
    const now = Date.now();
    if (_config && (now - _lastLoad < TTL_MS)) {
        return _config;
    }

    try {
        const res = await fetch(REMOTE_CONFIG_URL, {
            headers: { 'User-Agent': 'AnthologyLive/1.0' },
            signal: AbortSignal.timeout(5000)
        });
        if (res.ok) {
            _config = await res.json();
            _lastLoad = now;
            return _config;
        }
    } catch (e) {}

    const local = readLocalConfig();
    if (local) {
        _config = local;
        _lastLoad = now;
        return _config;
    }

    _config = _config || {};
    return _config;
}

function val(dotPath) {
    if (!_config) return null;
    const parts = dotPath.split('.');
    let cur = _config;
    for (const p of parts) {
        if (!cur || typeof cur !== 'object') return null;
        cur = cur[p];
    }
    return cur !== undefined ? cur : null;
}

module.exports = {
    loadConfig,
    val
};
