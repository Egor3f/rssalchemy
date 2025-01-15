function readSpecsForm() {
    let specs = {};
    for (let field of document.forms['wizard'].elements) {
        specs[field.name] = field.value;
    }
    return specs;
}

function writeSpecsToForm(specs) {
    for (let [k, v] of Object.entries(specs)) {
        document.forms['wizard'].elements[k].value = v;
    }
}

async function encodeSpecs(specs) {
    let byteArray = new TextEncoder().encode(JSON.stringify(specs));
    let cs = new CompressionStream('deflate-raw');
    let writer = cs.writable.getWriter();
    writer.write(byteArray);
    writer.close();
    let response = new Response(cs.readable);
    let respBuffer = await response.arrayBuffer();
    let b64str = btoa(String.fromCharCode.apply(null, new Uint8Array(respBuffer)));
    return b64str.replaceAll('=', '');
}

async function decodeSpecs(str) {
    const byteArray = Uint8Array.from(atob(str), c => c.charCodeAt(0));
    let ds = new DecompressionStream('deflate-raw');
    let writer = ds.writable.getWriter();
    writer.write(byteArray);
    writer.close();
    let response = new Response(ds.readable);
    let respText = await response.text();
    return JSON.parse(respText);
}

function displayUrl(url) {
    let link = document.getElementById('ready_url_link');
    link.href = url;
    link.style.visibility = 'visible';
    let readyUrlInput = document.getElementById('url_input');
    readyUrlInput.value = url;
    readyUrlInput.focus();
    readyUrlInput.select();
    document.getElementById('cont_url_len').innerText = `len=${url.length}`;
}

function baseUrl() {
    return document.location.origin + '/api/v1/render/';
}

async function genUrl() {
    let specs = readSpecsForm();
    let encodedSpecs = await encodeSpecs(specs);
    let url = baseUrl() + encodedSpecs;
    displayUrl(url);
}

async function editUrl() {
    let url = document.getElementById('url_input').value;
    let specs = await decodeSpecs(url.replace(baseUrl(), ''));
    writeSpecsToForm(specs);
    displayUrl(url);
}

document.addEventListener('DOMContentLoaded', ev => {
    document.getElementById('btn_gen_url').addEventListener('click', genUrl);
    document.getElementById('btn_edit').addEventListener('click', editUrl);
});
