"use strict";
function ajax_gen(url, data1, success, error, logical_failed = undefined, context=this, method="post"){
	return $.ajax({
		url: url,
		async: true,
		data: data1,
		dataType: "json",
		success: function(ret_data, textStatus, jqXHR){
			let ret = undefined;
			if(typeof(textStatus) === "string" && textStatus === "success"){
				if(typeof(ret_data) == "object" && ret_data.status === 0){
					if(typeof(success) == "function"){
						ret = success(ret_data, textStatus, jqXHR);
					}else{
						alert_msg($(this), "参数错误: success需为function");
						if(typeof(logical_failed) === "function") logical_failed();
					}
				}else{
					alert_msg($(this), "服务器未正常返回回执, 请联系管理员.<br>" + get_string(ret_data));
					if(typeof(logical_failed) === "function") logical_failed();
				}
			}else{
				alert_msg($(this), "本地执行器无法解析返回内容, 请联系管理员.<br>" + get_string(ret_data));
				if(typeof(logical_failed) === "function") logical_failed();
			}
			if(typeof(ret) === "function"){
				return ret();
			}
			return ret;
		},
		error: error,
		type: method,
		timeout: 5000,
		context: context,
	})
}
function set_enable(form_jd, bool){
	if(!bool){
		form_jd.find("button").addClass("disabled");
		form_jd.find("input,textarea").attr("readonly", true);
		form_jd.find(".back").addClass("d-none");
		form_jd.find(".spinner-border").removeClass("d-none");
		form_jd.find(".close").addClass("d-none");
	}else{
		form_jd.find("button").removeClass("disabled");
		form_jd.find("input,textarea").attr("readonly", null);
		form_jd.find(".back").removeClass("d-none");
		form_jd.find(".spinner-border").addClass("d-none");
		form_jd.find(".close").removeClass("d-none");
	}
}
function get_type(res_text){
	if(typeof(res_text) !== "string"){
		return "{Empty}";
	}
	let trimed = res_text.trim();
	if(trimed === ""){
		return "{Empty}"
	}
	if(trimed.startsWith("<!DOCTYPE") || trimed.startsWith("<html")){
		return "HTML";
	}else if(trimed.startsWith("<?")){
		if(trimed.substring(0, 20).indexOf("xml") != -1){
			return "XML";
		}else if(trimed.substring(0, 20).indexOf("php") != -1){
			return "PHP";
		}else{
			return "STRUCTED_SOURCE";
		}
	}else{
		try{
			$.parseJSON(trimed)
		}catch(e){
			return "UNKNOWN"
		}
		return "JSON"
	}
}
function get_string(res_text, length = 100){
	switch(get_type(res_text)){
		case "HTML":
			return "{HTML}: " + res_text.trim().substring(0, length).replace("<", "&lt;").replace(">", "&gt;");
		case "XML":
			return "{XML}: " + res_text.trim().substring(0, length).replace("<", "&lt;").replace(">", "&gt;");
		case "JSON":
			return res_text.trim().substring(0, length);
		case "{Empty}":
			return "{Empty}";
		case "UNKNOWN":
			return "{UNKNOWN}"
		default:
			return "{Unstructed string}";
	}
}
function alert_msg(form_jd, msg){
	if(msg.length > 0){
		form_jd.find(".help_text .alert").html(msg);
		form_jd.find(".help_text").removeClass("d-none");
	}else{
		form_jd.find(".help_text").addClass("d-none")
	}

}
function ajax_failed_handler(jqXHR, textStatus, errorThrown){
	alert_msg($(this), "服务器连接失败, 请联系管理员, 可提供以下信息.<br>" +
		"StatusCode: " + jqXHR.status + " " + jqXHR.statusText + "<br>" +
		"Response: " + get_string(jqXHR.responseText) + "");
	set_enable($(this), true);
}
