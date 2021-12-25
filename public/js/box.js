function toast(msg, afterHidden, beforeShow){
	$.toast({
	    text: '<div style="text-align:center;">'+msg+'</div>',
	    position: 'mid-center',
	    showHideTransition: 'fade',
	    stack: false,
	    hideAfter: 1000,
	    allowToastClose: false,
	    loader: false,
	    beforeShow: function () {
	    	if (typeof beforeShow == 'function') {
	    		beforeShow();
	    	} 	
	    }, // will be triggered before the toast is shown
    	afterShown: function () {
    	}, // will be triggered after the toat has been shown
    	beforeHide: function () {
    	}, // will be triggered before the toast gets hidden
    	afterHidden: function () {
    		if (typeof afterHidden == 'function') {
	    		afterHidden();
	    	}
    	}  // will be triggered after the toast has been hidden
	});
}

function selectAll(obj){
	$('input[name=mail_select]').attr("checked",obj.checked);
	var mailSelect = $('input[name=mail_select]');
	getSelectVal();
}


function getSelectVal(){
	var ids = "";
	$("input[name=mail_select]").each(function(){
	   if(this.checked){
	       console.log("id",$(this).val());
	   }
	});
}

function setMailStar(obj){
	var id = $(obj).attr("data-id");

	console.log(id,obj);

	toast("设置成功!");
}