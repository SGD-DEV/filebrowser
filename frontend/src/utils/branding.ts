// Uploaded logos and icons keep their URLs (logo.svg may really be a PNG), so
// a version from the server is appended to make browsers load a new upload
// instead of a cached older image.

export const brandingVersion = (): string =>
  window.FileBrowser.BrandingVersion || "0";

export const withBrandingVersion = (url: string, version = brandingVersion()) =>
  `${url.split("?")[0]}?v=${version}`;

export function applyBrandingVersion(version = brandingVersion()) {
  window.FileBrowser.BrandingVersion = version;

  document
    .querySelectorAll<HTMLLinkElement>(
      'link[rel~="icon"], link[rel="apple-touch-icon"]'
    )
    .forEach((link) => {
      // The file behind favicon.svg may be a PNG; a stale type would make
      // some browsers skip it.
      link.removeAttribute("type");
      link.href = withBrandingVersion(link.href, version);
    });

  document
    .querySelectorAll<HTMLImageElement>('img[src*="/img/logo."]')
    .forEach((img) => {
      img.src = withBrandingVersion(img.src, version);
    });
}
