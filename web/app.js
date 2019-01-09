


async function main () {
    let res = await fetch("/list")

    let data = await res.json()

    let html = ""

    for (let k in data) {
        let isGroup = data[k].Actions
        let img = ""
        if (data[k].Icon) {
            img = `<img src="${data[k].Icon}">`
        }
        html += `<a onclick="send('${k}'); return false;" href="${link(k)}" class="action ${isGroup ? 'group' : ''}">
            <input type="checkbox" value="${k}">
            ${img}
            <div>${k}</div>
        </a>`
    }

    let container = document.querySelector(".actions")

    container.innerHTML = html
}

function onLearnSubmit(el) {
   el.action = `/learn/${el.querySelector('.n').value}/${encodeURIComponent(el.querySelector('.i').value)}`
}

function onGroupSubmit (el) {
    actions = [...document.querySelectorAll(".action input:checked")]
    el.action = `/group/${el.querySelector('.n').value}/${encodeURIComponent(el.querySelector('.i').value)}/${actions.map(e => e.value).join('/')}`
}

function onDeleteSubmit(el) {
    let name = document.querySelector('.action input:checked').value
    el.action = `/delete/${name}`
}

function onRenameSubmit(el) {
    let from = document.querySelector('.action input:checked').value
    let to = el.querySelector("input").value
    el.action = `/rename/${from}/${to}`
}

async function send(name) {
    navigator.vibrate(100)

    new Noty({
        theme: 'relax',
        type: 'info',
        text: 'Do: ' + name,
        timeout: 500,
        layout: 'bottomRight'
    }).show();

    await fetch(link(name))

    new Noty({
        theme: 'relax',
        type: 'success',
        text: 'Action Done: ' + name,
        timeout: 1000,
        layout: 'bottomRight'
    }).show();
}

function link(name) {
    let token = encodeURIComponent(document.cookie)    
    return `/send/${name}?token=${token}`
}

main().catch((err) => {
    alert(err)
})