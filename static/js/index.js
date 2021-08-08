import 'bootstrap/scss/bootstrap.scss'

import 'prismjs';
import {render, html} from 'uhtml'
import Fuse from 'fuse.js'

const Search = (searchField, autocompleteBox) => {
     var objList = Object.entries(goIndex).map(x => ({key: x[0], text: x[1].replace(/(<([^>]+)>)/gi, "")}))
    const fuseOpts = {
        includeScore: true,
        keys: ['text'],
        includeMatches: true,
        threshold: 0.3,
        ignoreLocation: true,
    }
    const fuse = new Fuse(objList, fuseOpts)
    const {render, html} = uhtml;
    searchField.addEventListener("blur", () => {
        setTimeout(() => render(autocomplete, html``), 250)
    })
    const urlSearchParams = new URLSearchParams(window.location.search);
    const params = Object.fromEntries(urlSearchParams.entries());
    if (params.hil) {
        var arr = JSON.parse(params.hil).join(" ");
        var doc = [];
        const possibleNodes = [
            ...document.querySelectorAll('main .main-container>div>*:not(ul)'),
            ...document.querySelectorAll('main .main-container>div>ul>*')];
        possibleNodes.map(x => doc.push({
            node: x,
            value: x.innerText
        }))
        var nnode = new Fuse(doc, {keys: ['value'], ignoreLocation: true, threshold: 0.2}).search(arr)[0].item.node;
        nnode.id = '_found';
        window.location.hash = '#_found';
    }
    const renderComplete = (e) => {
        const query = e.target.value.toLowerCase();
        const outList = fuse.search(e.target.value).map(result => {
            let words = result.item.text.split(/\s+/)
            var match = new Fuse(words, {threshold: 0.5}).search(e.target.value);
            const idx = words.indexOf(match[0].item);
            var clamp = (n, low, high) => n > high ? high : n < low ? low : n;
            var fromIdx = clamp(idx - 8, 0, idx - 1);
            var toIdx = clamp(idx + 8, idx + 1, words.length);
            var outw = [];
            var wordsRaw = [];
            for (var i = fromIdx; i < toIdx; i++) {
                wordsRaw.push(words[i])
                if (i == idx)
                    outw.push(html`<strong>${words[i]} </strong>`)
                else
                    outw.push(html`<span>${words[i]} </span>`)
            }
            const cl = () => document.location = result.item.key + ".html?hil=" + encodeURIComponent(JSON.stringify(wordsRaw));
            return html`<label @click=${cl}>${outw}</label>`
        })
        render(autocomplete, html` ${outList}`);
    };
    searchField.addEventListener("input", renderComplete);
    searchField.addEventListener("focus", renderComplete);
	window.Prism = Prism;
};

export { Search };
