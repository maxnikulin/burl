/* Copyright (C) 2021 Max Nikulin */
/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

"use strict";

var gRpcId = 0;
function getId() {
	return gRpcId++;
}

/* This class does not follow recommendation from "Chrome developers"
 * pages to terminate connections as soon as possible
 * to allow browser to unload extension while it is not used
 * and as a result to reduce memory footprint.
 */
class JsonRpcClient {
	constructor(applicationId, { debug, idGenerator } = {}) {
		if (!browser.runtime.connectNative) {
			throw new Error('Addon has no "nativeMessaging" in permissions');
		}
		this.applicationId = applicationId;
		this.debug = debug;
		this.getId = idGenerator || getId;
		this.promiseMap = new Map();
		this.tag = `${applicationId} JsonRpcClient`;
		this.onDisconnect = this.disconnected.bind(this);
		this.onMessage = this.messageReceived.bind(this);
	};

	async sendMessage(method, arg) {
		const port = this.getPort();
		const id = this.getId();
		const promise = new Promise(
			(resolve, reject) => this.promiseMap.set(id, { resolve, reject}));
		// JSON-RPC arguments should be an array,
		// Go net/rpc package supports only one argument,
		// so an array containing a singe object is the least common denominator.
		const request = { id, method, params: [arg] };
		this.debug && console.debug("%s: post %o", this.tag, request);
		try {
			port.postMessage(request);
		} catch (e) {
			this.promiseMap.delete(id);
			throw e;
		}
		return promise;
	};

	// Lazy connect
	getPort() {
		if (this.port && !this.port.error) {
			return this.port
		}

		this.debug && console.debug("%s: connecting", this.tag);
		this.port = browser.runtime.connectNative(this.applicationId);
		this.port.onMessage.addListener(this.onMessage);
		this.port.onDisconnect.addListener(this.onDisconnect);
		return this.port;
	};

	disconnected(port) {
		this.port = null;
		this.debug && console.debug("%s: disconnected %s", this.tag, port && port.name);
		if (port && port.error) {
			console.error("%s: disconnect error %s", this.tag, port.error);
		}
		port.onMessage.removeListener(this.onMessage);
		port.onDisconnect.removeListener(this.onDisconnect);
		const error = port.error || new Error("disconnected");
		for (let v of this.promiseMap.values()) {
			v.reject(error);
		}
		this.promiseMap.clear();
	};

	messageReceived(response) {
		this.debug && console.debug("%s: received %o", this.tag, response);
		if (response == null) {
			console.error("%s: null message received", this.tag);
			return;
		}
		const id = response.id;
		if (id == null) {
			console.error("%s: no id in received message %s", this.tag, JSON.stringify(response));
			return;
		}
		const promiseObj = this.promiseMap.get(id);
		if (promiseObj == null) {
			console.error("%s: message with unknown id received '%s' %s", this.tag, id, JSON.stringify(response));
			return;
		}
		if (response.result) {
			promiseObj.resolve(response.result);
		} else if (response.error) {
			promiseObj.reject(new Error(response.error));
		} else {
			return promiseObj.reject(new Error("invalid RPC response " + JSON.stringify(response)));
		}
		this.promiseMap.delete(id);
	};
}
