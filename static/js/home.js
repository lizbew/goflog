$(function() {
	var blogUrl = "/blog/api/v1/posts";
	$.ajax(blogUrl, {
		type: 'POST',
		dataType: 'json',
		context: $('#blog-list'),
		success: function(data, textStatus, jqXHR) {
			if (data.post_num > 0) {
				this.empty();
				var containerEle = this;
				jQuery.each(data.posts, function(index, post){
					var t = '<li><a href="'+post.url+'" title="'+post.title+'">'+post.title+'</a></li>';
					containerEle.append(t);
				});
			} else {
				containerEle.append('<li>No Post Found</li>');
			}
		},
		error: function(jqXHR, textStatus, errorThrown) {
			$('#blog-list').html('<li>Error when load...</li>');
		}
	});
});
