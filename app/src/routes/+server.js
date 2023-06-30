import { json } from '@sveltejs/kit';
import { units } from "$lib/server/units"
import { formula } from "$lib/server/formula"

export async function GET({ event, url, params }) {
    const rowText = url.searchParams.get("r");

    console.log("+server, urlparam: ", rowText)

    if (rowText.length == 0) {
        return json({ r: "" })
    }

    const res = await getResult(rowText);

    console.log("+server, res: ", res)

    if (res.fid === 0 && res.tid === 0) {
        return json({ r: "" })
    }

    const fu = units.find(e => e.unit_id == res.fid)
    const tu = units.find(e => e.unit_id == res.tid)

    if (fu == undefined || tu == undefined) {
        return json({ r: "" })
    }

    if (fu.group_id != tu.group_id) {
        return json({ r: "" })
    }

    const form = formula.find(e => e.id == fu.formula_id)
    const calcRes = form.calculate(fu.ratio, tu.ratio, res.val)


    return json({ r: calcRes })
}

async function getResult(row) {
    const resp = await fetch(`http://sc_goapp:8080/calc?r=${encodeURIComponent(row)}`)
    const res = await resp.json()
    return res
}