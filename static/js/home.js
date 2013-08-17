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
					var t = '<li><a href="'+post.url+'" title="'+post.title+'">'+post.title+'</a></li>';
					containerEle.append(t);
				});
			} else {
				containerEle.append('<li>No Post Found</li>')
			}
		}
	});
});