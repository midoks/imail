
function selectAll(obj){
	$('input[name=mail_select]').attr("checked",obj.checked);
	var mailSelect = $('input[name=mail_select]');
	console.log($('input[name=mail_select]'));

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

	$('.tiny.modal').modal('show');
}