
function ExecuteQuery() {
	// console.log("ExecuteQuery()");
	var q = getQuery();
	// console.log("query = " + q);
	// sleep(2000);
	q = '/get?q=' + encodeURIComponent(q);
	document.location = q;
}

function SetQueryButtonText() {
	var q = getQuery();
	document.getElementById('query-index-input-button').value = 'query: ' + q;
}

function getQuery() {
	var p = document.getElementById('query-index-input-predicates').value;
	p = toQuotedArgs(p);
	var c = document.getElementById('query-index-input-concepts').value;
	c = toQuotedArgs(c);
	if (p.length == 0) {
		return '?({' + c + '})';
	} else {
		return '?(' + p + '({' + c + '}))'
	}
}

function toQuotedArgs(arg) {
	if (arg.length == 0) {
		return arg;
	} else if (arg == "*") {
		return arg;
	}
	var args = arg.split(/\s*,\s*/);
	var res = [];
	for (var i=0; i < args.length; i++) {
		if (args[i].length > 0) {
			res.push('"' + args[i] + '"');
		}
	}
	return res.join(",");
}

function sleep(millis) {
	var date = new Date();
	var curdate = null;
	do {curdate = new Date();}while(curdate - date < millis);
}
