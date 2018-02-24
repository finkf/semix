// semix.js
var semix={};

Storage.prototype.setObject = function(key, value) {
    this.setItem(key, JSON.stringify(value));
}

Storage.prototype.getObject = function(key) {
    var value = this.getItem(key);
    return value && JSON.parse(value);
}

function toQuotedArgs(arg) {
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

semix.LoadConfiguration = function() {
		var config = semix.getConfiguration();
		if (config.limits.length > 0) {
				document.getElementById('error-limits').value = config.limits.join(',');
		}
		if (config.resolvers.length > 0) {
				document.getElementById('resolvers').value = config.resolvers.join(',');
		}
};

semix.SaveConfiguration = function() {
		var config = {
				'resolvers': semix.getResolversFromElement(),
				'limits': semix.getErrorLimitsFromElement()
		};
		localStorage.setObject('semix', config);
		// reload Configuration on page
		semix.LoadConfiguration();
};

semix.getConfiguration = function() {
		var config = localStorage.getObject('semix');
		if (config === null) {
				config = {
						'limits': [],
						'resolvers': []
				};
		};
		return config;
};
