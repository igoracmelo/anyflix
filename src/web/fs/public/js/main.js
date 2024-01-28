/** @param {HTMLInputElement[]} inputs */
function queryObject(inputs) {
    let query = {}
    for (let input of inputs) {
        if (input.type === 'checkbox' && !input.checked) {
            continue
        }
        if (input.type === 'radio' && !input.selected) {
            continue
        }
        if (input.type == 'submit') {
            continue
        }
        query[input.name] = input.value
    }

    return query
}

/** @param {Record<string, string>} obj */
function queryString(obj) {
    let parts = []
    for ([key, val] of Object.entries(obj)) {
        parts.push(`${key}=${val}`)
    }
    return parts.join('&')
}

function updateQuery(queryString) {
    history.pushState(null, '', location.pathname + '?' + queryString)
}