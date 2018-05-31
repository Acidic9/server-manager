interface Array<T> {
	pushIfNotExist(e: T): T[];
	delete(e: T): T[];
}

interface JQuery<TElement extends Node = HTMLElement> extends Iterable<TElement> {
	select2: any;
	effect: any;
}