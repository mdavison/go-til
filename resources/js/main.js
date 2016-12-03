// Using an object literal for a jQuery feature
var til = {
    init: function() {
        $("#add-item-form").submit(function(event) {
            event.preventDefault();
            til.addItem();
        });

        $("ul#items").on("click", "a.delete", function(event){
            event.preventDefault();
            var id = $(this).attr("id").split('-')[1];
            console.log("the id is: " + id);
            til.deleteItem(id);
        })

        $('body').on('click', '[data-editable]', function(){
            til.editItem($(this));
        });

        $('div').on('click', '#show-register', function(){
            $('#register-form').removeClass('hidden');
            $('#show-register').addClass('active');
            $('#login-form').addClass('hidden');
            $('#show-login').removeClass('active');
            $('#show-register').blur();
            $('#login-email').focus();
        });

        $('div').on('click', '#show-login', function(){
            $('#register-form').addClass('hidden');
            $('#show-register').removeClass('active');
            $('#login-form').removeClass('hidden');
            $('#show-login').addClass('active');
            $('#show-login').blur();
            $('#register-email').focus();
        });

    },

    addItem: function() {
        $.ajax({
            url: "/create",
            method: "POST",
            data: $("#add-item-form").serialize(),
            success: function(rawData){
                var parsed = JSON.parse(rawData);
                if (!parsed) {
                    return false;
                }

                // Clear the form
                $('#title').val('');
                $('#msg').remove();

                // Add in the new items
                var haveUser = true
                parsed.forEach(function(result){
                    if (result.ID == -1) {
                        haveUser = false
                    }
                    $("ul#items").append('<li><strong>' + result.Date + '</strong> <span id="item-' + result.ID + '" data-editable>' + result.Title + '</span> <a href="/delete" class="delete pull-right" id="delete-' +  result.ID + '">Delete</a></li>');
                });
                if (!haveUser) {
                    $("ul#items").prepend('<p class="text-warning" id="msg">You must be logged in for items to be saved.</p>');
                }
            }
        });
    },

    deleteItem: function(id) {
        $.ajax({
            url: "/delete/" + id,
            method: "GET",
            success: function(){
                // Remove the item from the UI
                $("#delete-" + id).parent().remove();
            }
        });
    },

    editItem: function(item) {
        var input = $('<input/>').val( item.text());
        var id = item.attr("id").split('-')[1];
        item.replaceWith(input);

        input.keydown(function(event){
            if ( event.which == 13 ) {
                save();
            }
        });

        var save = function(){
            var editedText = input.val();
            var editedItemSpan = $('<span id="item-' + id + '" data-editable />').text(editedText);
            input.replaceWith(editedItemSpan);

            $.ajax({
                url: "/edit",
                method: "POST",
                data: JSON.stringify({"ID": id, "Title": editedText}),
                contentType: "application/json; charset=utf-8"
            });
        };

        input.one('blur', save).focus();
    }
};

$( document ).ready( til.init );