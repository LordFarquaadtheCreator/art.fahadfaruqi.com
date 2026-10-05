// Whether the viewer is up. Kept as shared state so the page can open it and chrome
// outside the page can stand down, without either knowing about the other.
export const viewer = $state({ open: false });
