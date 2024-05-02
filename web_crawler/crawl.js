import { JSDOM } from "jsdom";

function normalizeURL(url) {
    const urlObj = new URL(url);
    let fullPath = `${urlObj.host}${urlObj.pathname}`;
    if (fullPath.slice(-1) === "/") {
        fullPath = fullPath.slice(0, -1);
    }
    return fullPath;
}

function getURLsFromHTML(htmlBody, baseURL) {
    const urls = [];
    const dom = new JSDOM(htmlBody);
    const anchors = dom.window.document.querySelectorAll("a");

    for (const anchor of anchors) {
        if (anchor.hasAttribute("href")) {
            let href = anchor.getAttribute("href");

            try {
                // convert any relative URLs to absolute URLs
                href = new URL(href, baseURL).href;
                urls.push(href);
            } catch (err) {
                console.log(`${err.message}: ${href}`);
            }
        }
    }

    return urls;
}

async function fetchHTML(url) {
    let res;
    try {
        res = await fetch(url);
    } catch (err) {
        throw new Error(`Got Network error: ${err.message}`);
    }

    if (res.status > 399) {
        throw new Error(`Got HTTP error: ${res.status} ${res.statusText}`);
    }

    const contentType = res.headers.get("content-type");
    if (!contentType || !contentType.includes("text/html")) {
        throw new Error(`Got non-HTML response: ${contentType}`);
    }

    return res.text();
}

async function crawlPage(baseURL, currentURL = baseURL, pages = {}) {
    // if this is an offsite URL, bail immediately
    const currentURLObj = new URL(currentURL);
    const baseURLObj = new URL(baseURL);
    if (currentURLObj.hostname !== baseURLObj.hostname) {
        return pages;
    }

    // use a consistent URL format
    const normalizedURL = normalizeURL(currentURL);

    // if we've already visited this page
    // just increase the count and don't repeat
    // the http request
    if (pages[normalizedURL] > 0) {
        pages[normalizedURL]++;
        return pages;
    }

    // initialize this page in the map
    // since it doesn't exist yet
    pages[normalizedURL] = 1;

    // fetch and parse the html of the currentURL
    console.log(`crawling ${currentURL}`);
    let html = "";
    try {
        html = await fetchHTML(currentURL);
    } catch (err) {
        console.log(`${err.message}`);
        return pages;
    }

    // recur through the page's links
    const nextURLs = getURLsFromHTML(html, baseURL);
    for (const nextURL of nextURLs) {
        pages = await crawlPage(baseURL, nextURL, pages);
    }

    return pages;
}

export { normalizeURL, getURLsFromHTML, crawlPage };
