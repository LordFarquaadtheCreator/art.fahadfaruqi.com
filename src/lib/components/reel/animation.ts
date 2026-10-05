// The contract between the engine and a pinned animation.
//
// The engine owns the pin and the four phases — entry, work, hold, exit — and hands the
// animation one state object per tick. The animation owns everything else: how long its
// work phase runs, and what happens to its marks while it does.
//
// The Turn is the only animation on the page today: it treats the section slots as a
// sequence and writes each one its --c and --gate. But nothing here knows that — a new
// pinned mechanic is a new animation object, not an engine change.

export interface Mark {
	el: HTMLElement;
	/** its position in the turn, for the counter */
	index: number;
}

export interface ReelContext {
	/** the pinned element */
	node: HTMLElement;
	/** every [data-reel-frame] under it, in DOM order */
	marks: Mark[];
	/** current viewport height */
	vh: number;
}

export interface ReelTick {
	/** the crank, 0..1 across the work phase */
	p: number;
	/** entry envelope, 0..1 — the section arriving */
	entry: number;
	/** exit envelope, 0..1 — the section handing off to the next chapter */
	exit: number;
	vh: number;
	marks: Mark[];
	/** writes a global frame number into the counter */
	show: (index: number) => void;
}

export interface ReelAnimation {
	/**
	 * Called on mount and on every resize. Returns the work phase's scroll length in px,
	 * and may cache whatever geometry the tick needs.
	 */
	measure: (context: ReelContext) => number;
	/** Called once per frame while the section is pinned and near the viewport. */
	tick: (state: ReelTick) => void;
}
