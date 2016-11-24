// Using an object literal for a jQuery feature
var til = {
    init: function() {
        $("#add-item-form").submit(function(event) {
            event.preventDefault();
            til.addItem();
        });

        // $("a.delete").click(function(event){
        //     event.preventDefault();
        //     //console.log("delete " + $(this).attr("id"));
        //     var id = $(this).attr("id").split('-')[1];
        //     console.log("the id is: " + id);
        //     til.deleteItem(id);
        // });
        $("ul#items").on("click", "a.delete", function(event){
            event.preventDefault();
            var id = $(this).attr("id").split('-')[1];
            console.log("the id is: " + id);
            til.deleteItem(id);
        })
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
                    $("ul#items").append("<li>" + result.Title + ' <a href="/delete" class="delete" id="delete-' +  result.ID + '">Delete</a></li>');
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
    }
};

$( document ).ready( til.init );