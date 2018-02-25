// semix.js
var semix={};
semix.config = {
		'limits': [],
		'resolvers': [],
		'threshold': 0.5,
		'memorySize': 10
};

// setObject stores JSON object as string.
Storage.prototype.setObject = function(key, value) {
		this.setItem(key, JSON.stringify(value));
}

// getObject returns a JSON encoded object.
Storage.prototype.getObject = function(key) {
    var value = this.getItem(key);
    return value && JSON.parse(value);
}

function toQuotedArgs(arg) {eor
		if (arg.length === 0) {
				return arg;
		}
		if (arg === "*") {
				return arg;
		}
		var args = arg.split(/\s*,\s*/);
		var res = [];
		var i;
		for (i=0; i < args.length; i++) {
				if (args[i].length > 0) {
						if (args[i].substring(0, 1) === "!") {
								res.push('!"' + args[i].substring(1) + '"');
						} else {
								res.push('"' + args[i] + '"');
						}
				}
		}
		return res.join(",");
}

function getQuery() {
		var p = document.getElementById('query-index-input-predicates').value;
		p = toQuotedArgs(p);
		var c = document.getElementById('query-index-input-concepts').value;
		c = toQuotedArgs(c);
		if (p.length === 0) {
				return '?(' + c + ')';
		}
		return '?(' + p + '(' + c + '))';
}

function sleep(millis) {
		var date = new Date();
		var curdate = null;
		do {curdate = new Date();}while(curdate - date < millis);
}

semix.SetQueryButtonText = function() {
		var q = getQuery();
		document.getElementById('query-index-input-button').value = 'query: ' + q;
}

semix.ExecuteQuery = function() {
		var q = getQuery();
		q = '/get?q=' + encodeURIComponent(q) + '&n=50&s=0';
		document.location = q;
};

semix.getErrorLimitsFromElement = function() {
		var e = document.getElementById('error-limits');
		if (e === null) {
				return [];
		}
		var ks = e.value.split(',');
		if (ks === null) {
				return [];
		}
		var res = [];
		ks.forEach(function(e, i) {
				var n = parseInt(e, 10);
				if (isNaN(n) || n < 1 || n > 10) {
						return;
				}
				res.push(n);
		});
		res.sort(function(a,b){return a-b;});
		return res;
};

semix.getResolversFromElement = function() {
		var e = document.getElementById('resolvers');
		if (e === null) {
				return [];
		}
		var rs = e.value.split(',');
		if (rs === null) {
				return [];
		}
		var res = [];
		rs.forEach(function(e, i) {
				if (e === "") {
						return;
				}
				res.push(e);
		});
		return res;
};

semix.getThresholdFromElement = function() {
		var e = document.getElementById('threshold');
		if (e === null) {
				semix.config.threshold;
		}
		var n = parseFloat(e.value);
		if (isNaN(n) || n < 0 || n > 1) {
				semix.config.threshold;
		}
		return n;
};

semix.getMemorySizeFromElement = function() {
		var e = document.getElementById('memory-size');
		if (e === null) {
				return semix.config.memorySize;
		}
		var n = parseFloat(e.value);
		if (isNaN(n) || n < 0) {
				return semix.config.memorySize;
		}
		return n;
};

semix.LoadConfiguration = function() {
		var config = semix.getConfiguration();
		if (config.limits.length > 0) {
				document.getElementById('error-limits').value = config.limits.join(',');
		}
		if (config.resolvers.length > 0) {
				document.getElementById('resolvers').value = config.resolvers.join(',');
		}
		document.getElementById('threshold').value = config.threshold;
		document.getElementById('memory-size').value = config.memorySize;
};

semix.SaveConfiguration = function() {
		var config = {
				'resolvers': semix.getResolversFromElement(),
				'limits': semix.getErrorLimitsFromElement(),
				'threshold': semix.getThresholdFromElement(),
				'memorySize': semix.getMemorySizeFromElement(),
		};
		localStorage.setObject('semix', config);
		// Reload Configuration on page.
		// This way a valid configuration is inserted
		// in all input fields.
		semix.LoadConfiguration();
};

semix.getConfiguration = function() {
		var config = localStorage.getObject('semix');
		if (config === null) {
				return semix.config;
 		}
		return config;
};

semix.makeResolvers = function(config) {
		res = [];
		config.resolvers.forEach(function(e, i) {
				res.push(semix.makeResolver(e, config));
		});
		return res;
};

semix.makeResolver = function(name, config) {
		return {
				'Name': name,
				'MemorySize': config.memorySize,
				'Threshold': config.Threshold
		};
};

semix.PutContent = function() {
		var e = document.getElementById('input-content');
		if (e === null) {
				return;
		}
		semix.put(e.value, false);
};

semix.put = function(content, isURL) {
		var config = semix.getConfiguration();
		var data = {
				"Errors": config.limits,
				"Resolvers": semix.makeResolvers(config)
		};
		if (isURL === true) {
				data.ContentType = "text/html";
				data.URL = content;
		} else {
				data.ContentType = "text/plain";
				data.Content = content;
		}
		console.log("[POST] " + JSON.stringify(data));
		// At this point the post data is complete
		var post = new XMLHttpRequest();
		post.open("POST", "/put", true);
		post.onreadystatechange = function() {
				if (this.readyState === 4) {
						if (this.status !== 200) {
								semix.alertError(this.responseText, this.status);
								return;
						}
						document.write(this.responseText);
				}
		};
		post.setRequestHeader(
				"Content-type",
				"application/json; charset=UTF-8");
		post.send(JSON.stringify(data));
};

semix.alertError = function(msg, status) {
		window.alert("error:\n" + msg + "\nstatus: " + status);
};

// semix.getPutURL = function(base, url) {
// 		var config = semix.getConfiguration();
// 		var uri = base + '?m=' + config.memorySize;
// 		uri += '&t=' + config.threshold;
// 		if (url !== undefined) {
// 				base += '&url=' + url;
// 		}
// 		config.limits.forEach(function(e, i) {
// 				uri += '&ks=' + e;
// 		});
// 		config.resolvers.forEach(function(e, i) {
// 				uri += '&rs=' + e;
// 		});
// }
