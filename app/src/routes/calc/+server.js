import { json } from '@sveltejs/kit';

export async function GET({ event, url, params }) {
    const rowText = url.searchParams.get("r");
    const res = await getResult(rowText);
    const r = await reqm();

    return json({ val: res.val, fid: res.fid, tid: res.tid, rqm: r})
}

async function getResult(row) {
    const resp = await fetch(`http://sc_goapp:8080/calc?r=${encodeURIComponent(row)}`)
    const res = await resp.json()
    return res
}

async function reqm() {
    const resp = await fetch(`http://sc_nodemongoapi:3001/`)
    return resp.status
}