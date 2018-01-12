
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

function SetQueryButtonText() {
	var q = getQuery();
	document.getElementById('query-index-input-button').value = 'query: ' + q;
}

function ExecuteQuery() {
	var q = getQuery();
	q = '/get?q=' + encodeURIComponent(q) + '&n=50&s=0';
	document.location = q;
}
