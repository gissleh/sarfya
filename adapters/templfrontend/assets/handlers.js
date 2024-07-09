function search(el) {
    const filter = encodeURIComponent(el.querySelector("input.search-box").value);

    if (!!filter) {
        window.location.href = "/search/" + filter;
    } else {
        window.location.href = "/";
    }

    return false;
}

let hoverExampleId = "";
let hoverWordIds = [];

function onHover(el) {
    const ids = JSON.parse(el.dataset.ids);
    const exampleId = el.parentNode.dataset.id;

    if (!!hoverExampleId) {
        const prev = document.querySelector("#example-"+hoverExampleId).querySelectorAll("span, a");
        for (const el of prev) {
            el.classList.remove("hover");
        }
    }

    const current = document.querySelector("#example-"+exampleId).querySelectorAll("span, a");
    for (const partEl of current) {
        if (!partEl.dataset.ids) {
            continue;
        }

        const ids2 = JSON.parse(partEl.dataset.ids);
        if (ids2.find(id2 => ids.find(id => id === id2))) {
            partEl.classList.add("hover");
        }
    }

    hoverExampleId = exampleId;
    hoverWordIds = ids;
}

function onHoverEnd(el) {
    const ids = JSON.parse(el.dataset.ids);
    const exampleId = el.parentNode.dataset.id;

    if (hoverExampleId === exampleId && JSON.stringify(ids) === JSON.stringify(hoverWordIds)) {
        const prev = document.querySelector("#example-"+hoverExampleId).querySelectorAll("span, a");
        for (const el of prev) {
            el.classList.remove("hover");
        }
    }
}

// This is so that the textbox changes back if you go back after searching.
window.addEventListener("DOMContentLoaded", function() {
    const filter = decodeURIComponent(window.location.href.split("/").pop())
    const searchBox = document.querySelector("input.search-box");
    searchBox.value = filter;
});