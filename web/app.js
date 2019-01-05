


;(async() => {
    let res = await fetch("/list")

    let data = await res.json()

    let html = ""
    let token = encodeURIComponent(document.cookie)

    for (let k in data) {
        let actions = data[k].Actions ? `(${data[k].Actions})` : ''
        html += `<div class="action">
            <input type="checkbox" value="${k}">
            <a href="/send/${k}?token=${token}">${k} ${actions}</a>
        </div>`
    }

    let container = document.querySelector(".actions")

    container.innerHTML = html
    

})().catch((err) => {
    alert(err)
})

function onLearnSubmit(el) {
   el.action = `/learn/${el.querySelector('input').value}`
}

function onGroupSubmit (el) {
    let name = el.querySelector("input").value
    actions = [...document.querySelectorAll(".action input:checked")]
    el.action = `/group/${name}/${actions.map(e => e.value).join('/')}`
}

function onDeleteSubmit(el) {
    let name = document.querySelector('.action input:checked').value
    el.action = `/delete/${name}`
}