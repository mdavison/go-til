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

        // $("ul#items").on("click", "li", function(event){
        //     var id = $(this).attr("id").split('-')[1];
        //     console.log("edit item with id: " + id);
        //     til.editItem(id);
        // })

        $('body').on('click', '[data-editable]', function(){
            til.editItem($(this));
        });
    },

    addItem: function() {
        console.log("added item");
        $.ajax({
            url: "/create",
            method: "POST",
            data: $("#add-item-form").serialize(),
            success: function(rawData){
                console.log("success function called");
                var parsed = JSON.parse(rawData);
                console.log("parsed " + parsed);
                if (!parsed) {
                    return false;
                }

                // Clear the form
                $('#title').val('');

                // Add in the new items
                parsed.forEach(function(result){
                    $("ul#items").append("<li>" + result.Title + ' <a href="/delete" class="delete pull-right" id="delete-' +  result.ID + '">Delete</a></li>');
                });
            }
        });
    },

    deleteItem: function(id) {
        console.log("deleting item with id: " + id);
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
        console.log("editItem function called");

        var input = $('<input/>').val( item.text());
        var id = item.attr("id").split('-')[1];
        console.log("edit item with id of: " + id);
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
            console.log("edited text: " + editedText);

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