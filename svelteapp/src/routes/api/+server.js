import { json } from '@sveltejs/kit';
import { formulasList } from "$lib/server/formulas"


export async function GET({ event, url, params }) {
    const emptyResultObject = {
        result: ""
    }
    
    const rowToParse = url.searchParams.get("r");

    if (rowToParse.length <= 6) {
        return json(emptyResultObject)
    }

    rowToParse = rowToParse.substring(0, 70)

    const rowParsingResponse = await getResult(rowToParse);

    if (rowParsingResponse.status === 0) {
        return json(emptyResultObject)
    }
    if (rowParsingResponse.value === 0) {
        const resultObject = returnClonedObject(emptyResultObject)
        resultObject.result = 0
        return json(resultObject)
    }

    const formula = formulasList[rowParsingResponse.fid-1]
    const result = formula.calculate(rowParsingResponse.fur, rowParsingResponse.tur, rowParsingResponse.val)
    const resultObject = returnClonedObject(emptyResultObject)
    resultObject.result = result

    return json(resultObject)
}

async function getResult(row) {
    const resp = await fetch(`http://sc_goapp:8080/calc?r=${encodeURIComponent(row)}`)
    const res = await resp.json()
    return res
}
// async function t() {
//     const resp = await fetch(`http://sc_nodemongoapi:3001/?fugid=${encodeURIComponent(100)}&tugid=${encodeURIComponent(200)}&value=${encodeURIComponent(99)}`)
//     const res = await resp.json()
//     return res
// }

function returnClonedObject(objectToClone) {
    return JSON.parse(JSON.stringify(objectToClone))
}