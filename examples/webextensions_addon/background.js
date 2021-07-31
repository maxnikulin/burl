/* Copyright (C) 2021 Max Nikulin */
/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

"use strict";


const MENU_SQRT2 = "1";
const MENU_SQRT_M2 = "2";
const MENU_SLEEP = "3";

class Backend {
	constructor() {
		this.client = new JsonRpcClient("io.github.maxnikulin.burl_webextensions_example", { debug: true });
	};

	async sqrt(x) {
		console.log("Backend.sqrt(%o)", x);
		return this.client.sendMessage("example.Sqrt", x);
	};

	async sleep(ms) {
		console.log("Backend.sleep(%o)", ms);
		return this.client.sendMessage("example.Sleep", ms);
	};
}

function reportError(e) {
	// Sometimes error object is destroyed too quickly and unavailable
	// for inspection. Report string representation as fallback.
	console.error("error: %s %s %o", String(e), Object.prototype.toString.apply(e), e);
	if (e && e.stack) {
		console.log(e.stack);
	} else if (!(e instanceof Error)) {
		console.log(JSON.stringify(e));
	}
}

function onMenu({menuItemId}) {
	var p;
	switch (menuItemId) {
	case MENU_SQRT2:
		p = backend.sqrt(2);
		break;
	case MENU_SQRT_M2:
		p = backend.sqrt(-2);
		break;
	case MENU_SLEEP:
		p = backend.sleep(1500);
		break;
	default:
		p = Promise.reject(new Error("Unknown menuItemId " + menuItemId));
	}
	p.then(r => console.log("result: %o", r), reportError);
}

function createMenu() {
	browser.menus.onClicked.addListener(onMenu);
	const contexts = ["all"];
	function add(id, title) {
		browser.menus.create({id, contexts, title});
	}
	add(MENU_SQRT2, "sqrt(2)");
	add(MENU_SQRT_M2, "sqrt(-2)");
	add(MENU_SLEEP, "sleep()");
}

var backend;

function main() {
	browser.runtime.onInstalled.addListener(createMenu);
	backend = new Backend();
}

main();
