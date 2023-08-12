function updateDeployLink() {
	var link = document.getElementById("download-link")
	var target = document.getElementById("deploy-target")
	var demo = document.getElementById("deploy-demo")
	var http = document.getElementById("deploy-http")
	var currentURL = window.location.href
	var lastIndex = currentURL.lastIndexOf('/');
	var baseURL = currentURL.substring(0, lastIndex);
	var linkURL = "/deploy?target=" + target.value +
		"&demo=" + demo.checked +
		"&http=" + http.checked
	var downloadURL = baseURL + linkURL
	link.innerHTML = downloadURL
}

function stageDeploy() {
	var selectTarget = document.getElementById("deploy-target")
	var checkboxDemo = document.getElementById("deploy-demo")
	var checkboxHttp = document.getElementById("deploy-http")
	var downloadLink = document.getElementById("download-link")
	selectTarget.onchange = function(){updateDeployLink()}
	checkboxDemo.onchange = function(){updateDeployLink()}
	checkboxHttp.onchange = function(){updateDeployLink()}
	downloadLink.addEventListener("click", downloadFile)
	updateDeployLink()
}
