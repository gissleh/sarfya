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
    const extraIds = [];
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
            for (const id of ids2) {
                if (!extraIds.includes(id)) {
                    extraIds.push(id);
                }
            }

            partEl.classList.add("hover");
        }
    }

    for (const partEl of current) {
        if (!partEl.dataset.ids) {
            continue;
        }

        const ids2 = JSON.parse(partEl.dataset.ids);
        if (ids2.find(id2 => extraIds.find(id => id === id2))) {
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
    // Fill the search box with the filter, so that going back will replace it.
    let filter = decodeURIComponent(window.location.pathname.split("/").pop())
    if (!!window.location.searchParams && window.location.searchParams.has("q")) {
        filter = decodeURIComponent(window.location.searchParams.get("q"));
    }
    const searchBox = document.querySelector("input.search-box");
    searchBox.value = filter;

    // Find all hover-ables.
    const allHoverables = document.querySelectorAll("div.sentence span, div.sentence a");
    let list = [];
    for (const el of allHoverables) {
        if (!el.dataset.ids) {
            continue;
        }

        list.push(el);
    }

    // Process them in batches to leave room for other things to run.
    // This is only a problem on evil searches like "*"
    const handleBatch = function() {
        const current = list.slice(0, 128);
        list = list.slice(current.length);

        for (const el of current) {
            if (!el.dataset.ids) {
                continue;
            }

            el.onmouseenter = function() { onHover(el) };
            el.onmouseleave = function() { onHoverEnd(el) };
        }

        if (list.length > 0) {
            setTimeout(handleBatch, 10);
        }
    }
    setTimeout(handleBatch, 0);
});