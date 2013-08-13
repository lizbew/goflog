$(function() {
	var blogUrl = "/blog/api/v1/posts";
	$.ajax({
		url: blogUrl,
		context: $('#blog-list'),
		complete: function(jqXHR, textStatus) {
			var resp = jQuery.parseJSON(jqXHR.responseText);
			if (resp.post_num > 0) {
				this.empty();
				var containerEle = this;
				jQuery.each(resp.posts, function(index, post){
					var a = $('<li></li>').append(post.title);
					containerEle.append(a);
				});
			} else {
				containerEle.append('<li>No Post Found</li>')
			}
		}
	});
});