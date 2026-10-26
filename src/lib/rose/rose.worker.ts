import { buildRose } from './build';

// The app's TS program has the DOM lib, not lib.webworker, so the worker scope is described here.
type WorkerScope = { onmessage: (event: unknown) => void; postMessage: (value: unknown) => void };

const scope = self as unknown as WorkerScope;

// Roughly two thirds of a second of arithmetic: on the main thread that would be a visible freeze,
// here it is off the critical path and the page never waits for the rose.
scope.onmessage = () => {
	scope.postMessage(buildRose());
};
