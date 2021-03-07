/// <reference path="../../node_modules/@types/zepto/index.d.ts" />

// contains a list of free open trackers
const trackerListUrl = "https://cdn.jsdelivr.net/gh/ngosang/trackerslist/trackers_best.txt";

// fallbacks, grabbed from the tracker list url
const fallbackTrackers = [
    "udp://tracker.opentrackr.org:1337/announce",
    "udp://tracker.cyberia.is:6969/announce",
    "udp://exodus.desync.com:6969/announce",
    "udp://opentracker.i2p.rocks:6969/announce",
    "udp://47.ip-51-68-199.eu:6969/announce",
    "udp://tracker2.itzmx.com:6961/announce",
    "http://open.acgnxtracker.com:80/announce",
    "udp://tracker.tiny-vps.com:6969/announce",
    "udp://open.stealth.si:80/announce",
    "udp://www.torrent.eu.org:451/announce",
    "udp://tracker.torrent.eu.org:451/announce",
    "udp://retracker.lanta-net.ru:2710/announce",
    "udp://tracker.moeking.me:6969/announce",
    "udp://ipv4.tracker.harry.lu:80/announce",
    "udp://valakas.rollo.dnsabr.com:2710/announce",
    "udp://opentor.org:2710/announce",
    "http://tracker.dler.org:6969/announce",
    "udp://tracker.zerobytes.xyz:1337/announce",
    "udp://tracker.v6speed.org:6969/announce",
    "udp://tracker.uw0.xyz:6969/announce",
];

let trackers: string[];

$.ajax({
    type: 'GET',
    url: trackerListUrl,
    timeout: 300,
    success: function (data: string) {
        trackers = data.split("\n\n");
    },
    error: function () {
        trackers = fallbackTrackers;
    }
});

export function generateLink(infohash: string) {
    let link = `magnet:?xt=urn:btih:${infohash}`;
    if (trackers && trackers.length > 0) {
        link += "?";
        for (const tracker of trackers) {
            link += `tr=${encodeURIComponent(tracker)}&`;
        }
        link = link.substr(0, link.length-1);
    }
    return link;
}