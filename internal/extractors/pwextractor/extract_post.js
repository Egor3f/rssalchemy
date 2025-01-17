// let fnc = // for autocomplete
el => {
    let content = "";
    let paragraph = "";

    const finishParagraph = () => {
        content += "<p>" + paragraph + "</p>";
        paragraph = "";
    }

    const addImage = img => {
        let imgSrc = img.getAttribute('src');
        if (imgSrc.startsWith('/')) {
            imgSrc = `${document.location.origin}/${imgSrc}`;
        }
        content += `<img src="${imgSrc}"/>`;
    };

    let traverse = (node) => {
        // node = document.getRootNode(); // for autocomplete

        if (node.childNodes.length === 0) {
            return
        }

        for (let child of node.childNodes) {
            switch (child.nodeType) {
                case child.ELEMENT_NODE:
                    // child = document.getElementById(''); // for autocomplete

                    let tag = child.tagName.toLowerCase();

                    const allowedMarkupTags = ['b', 'i', 'strong'];
                    if (allowedMarkupTags.includes(tag)) {
                        paragraph += `<${tag}>`
                    }

                    if (tag === 'img') {
                        finishParagraph();
                        addImage(child);
                        break;
                    }

                    traverse(child);

                    if (allowedMarkupTags.includes(tag)) {
                        paragraph += `</${tag}>`
                    }

                    break;
                case child.TEXT_NODE:
                    if (child.nodeValue.length > 0) {
                        paragraph += child.nodeValue + " ";
                    }
                    break;
            }
        }
    };

    traverse(el);
    return content;
}
