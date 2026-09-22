// The label the pointer carries over the grid. Kept as shared state so a plate can set
// it and the atmosphere layer can draw it, without either knowing about the other.
export const cursor = $state({ label: '', active: false });
