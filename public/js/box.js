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
}


function getSelectVal(){
	var ids = "";
	$("input[name=mail_select]").each(function(){
	   if(this.checked){
	       ids += $(this).val()+',';
	   }
	});
	ids = $.trim(ids);
	if (ids.length>0){
		ids = ids.substring(0,ids.length-1);
	}
	return ids;
}

// set mail star
function setMailStar(obj){
	var id = $(obj).attr("data-id");
	var isHadStar = $(obj).hasClass('outline');

	if (isHadStar){
		$.post("/api/mail/star",{'ids':id}, function(data){
			toast(data['msg'],function(){
				location.reload();
			});
		});
	} else {
		$.post("/api/mail/unstar",{'ids':id}, function(data){
			toast(data['msg'],function(){
				location.reload();
			});
		});
	}
}

// move dir
function mailMove(dir){
	var ids = getSelectVal();
	if (ids.length==0){
		toast("no selected options");
		return;
	}

	$.post("/api/mail/move",{'ids':ids,"dir":dir}, function(data){
		toast(data['msg'],function(){
			location.reload();
		});
	});
}


function mailRead(val){
	var ids = getSelectVal();
	if (ids.length==0){
		toast("no selected options");
		return;
	}

	if (val>0) {
		$.post("/api/mail/read",{'ids':ids}, function(data){
			toast(data['msg'],function(){
				location.reload();
			});
		});
	} else {
		$.post("/api/mail/unread",{'ids':ids}, function(data){
			toast(data['msg'],function(){
				location.reload();
			});
		});
	}
}

function mailDeleted(val){
	var ids = getSelectVal();
	if (ids.length==0){
		toast("no selected options");
		return;
	}

	$.post("/api/mail/deleted",{'ids':ids}, function(data){
		toast(data['msg'],function(){
			location.reload();
		});
	});
}

//dd
function mailHardDeleted(val){
	var ids = getSelectVal();
	if (ids.length==0){
		toast("no selected options");
		return;
	}

	$.post("/api/mail/hard_deleted",{'ids':ids}, function(data){
		toast(data['msg'],function(){
			location.reload();
		});
	});
}


function mailExport(){
	var ids = getSelectVal();
	if (ids.length==0){
		toast("no selected options");
		return;
	}
	var url = '/mail/content/'+ids+'/download';
	window.open(url);
}