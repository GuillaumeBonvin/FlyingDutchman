// Reverse any String object
function reverseString(str) {
    var splitString = str.split(""); // var splitString = "hello".split("");
    var reverseArray = splitString.reverse(); // var reverseArray = ["h", "e", "l", "l", "o"].reverse();
    var joinArray = reverseArray.join(""); // var joinArray = ["o", "l", "l", "e", "h"].join("");

    return joinArray; // "olleh"
}

//require our websocket library
var WebSocketServer = require('ws').Server;

//creating a websocket server at port 9090
var wss = new WebSocketServer({port: 9090});

//all connected to the server users
var users = {};

//when a user connects to our sever
wss.on('connection', function (connection) {

    console.log("User connected");

    //when server gets a message from a connected user
    connection.on('message', function (message) {

        var data;
        //accepting only JSON messages
        try {
            data = JSON.parse(message);
        } catch (e) {
            console.log("Invalid JSON");
            data = {};
        }

        //switching type of the user message
        switch (data.Type) {
            //when a user tries to login

            case "login":
                console.log("User logged", data.Name);

                //if anyone is logged in with this username then refuse
                if (users[data.Name]) {
                    sendTo(connection, {
                        Type: "login",
                        Success: false
                    });
                } else {
                    //save user connection on the server
                    users[(data.Name)] = connection;
                    connection.Name = data.Name;

                    sendTo(connection, {
                        Type: "login",
                        Success: true
                    });

                    //if UserB exists then link connections A and B
                    var conn = users[reverseString(data.Name)];
                    if (conn != null) {
                        //setting that UserA connected with UserB
                        connection.otherName = reverseString(data.Name);
                        conn.otherName = data.Name;
                        console.log("Users linked", connection.Name + connection.otherName);

                        sendTo(conn, {
                            Type: "linked",
                            Offer: data.Offer,
                            Name: data.Sender
                        });
                        sendTo(connection, {
                            Type: "linked",
                            Offer: data.Offer,
                            Name: data.Sender
                        });
                    }
                }

                break;

            default:
                //any message sent by A is transferred to B
                var conn = users[connection.otherName];
                sendTo(conn, data);
                console.log("Sending message to: ", conn.Name);

                break;

        }
    });

    //when user exits, for example closes a browser window
    //this may help if we are still in "offer","answer" or "candidate" state
    connection.on("close", function () {

        if (connection.Name) {
            delete users[connection.Name];
            console.log("Disconnecting from ", connection.Name);


            if (connection.otherName) {
                console.log("Notifying linked user");
                var conn = users[connection.otherName];
                conn.otherName = null;

                if (conn != null) {
                    sendTo(conn, {
                        Type: "leave"
                    });
                }
            }
        }
    });


});

function sendTo(connection, message) {
    connection.send(JSON.stringify(message));
}


/*

//require our websocket library
var WebSocketServer = require('ws').Server;

//creating a websocket server at port 9090
var wss = new WebSocketServer({port: 9090});

//all connected to the server users
var users = {};

//when a user connects to our sever
wss.on('connection', function(connection) {

    console.log("User connected");

    //when server gets a message from a connected user
    connection.on('message', function(message) {

        var data;
        //accepting only JSON messages
        try {
            data = JSON.parse(message);
        } catch (e) {
            console.log("Invalid JSON");
            data = {};
        }

        //switching type of the user message
        switch (data.Type) {
            //when a user tries to login

            case "login":
                console.log("User logged", data.Name);

                //if anyone is logged in with this username then refuse
                if(users[data.Name]) {
                    sendTo(connection, {
                        Type: "login",
                        Success: false
                    });
                } else {
                    //save user connection on the server
                    users[(data.Name)] = connection;
                    connection.Name = data.Name;
                    console.log("connection.name is ", connection.Name, connection);



                    sendTo(connection, {
                        Type: "login",
                        Success: true
                    });
                }

                break;

            case "offer":
                //for ex. UserA wants to call UserB
                console.log("Sending offer to: ", data.Name);

                //if UserB exists then send him offer details
                var conn = users[data.Name];

                if(conn != null) {
                    //setting that UserA connected with UserB
                    connection.otherName = data.Name;

                    sendTo(conn, {
                        Type: "offer",
                        Offer: data.Offer,
                        Name: data.Sender
                    });
                } else {
                    sendTo(connection, {
                        Type: "noMatch",
                    })
                }

                break;

            case "reject":
                console.log("Sending reject to: ", data.Name)
                var conn = users[data.Name];

                if(conn != null) {
                    connection.otherName = data.Name;
                    sendTo(conn, {
                        Type: "reject",
                    });
                }

                break;
            case "answer":
                console.log("Sending answer to: ", data.Name);
                //for ex. UserB answers UserA
                var conn = users[data.Name];

                if(conn != null) {
                    connection.otherName = data.Name;
                    sendTo(conn, {
                        Type: "answer",
                        Answer: data.Answer
                    });
                }

                break;

            case "candidate":
                console.log("Sending candidate to:",data.Name);
                var conn = users[data.Name];

                if(conn != null) {
                    sendTo(conn, {
                        Type: "Candidate",
                        Candidate: data.Candidate
                    });
                }

                break;

            case "leave":
                console.log("Disconnecting from", data.Name);
                var conn = users[data.Name];
                conn.otherName = null;

                //notify the other user so he can disconnect his peer connection
                if(conn != null) {
                    sendTo(conn, {
                        Type: "leave"
                    });
                }

                break;

            default:
                sendTo(connection, {
                    Type: "error",
                    Message: "Command not found: " + data.Type
                });

                break;
        }
    });

    //when user exits, for example closes a browser window
    //this may help if we are still in "offer","answer" or "candidate" state
    connection.on("close", function() {

        if(connection.Name) {
            delete users[connection.Name];

            if(connection.otherName) {
                console.log("Disconnecting from ", connection.otherName);
                var conn = users[connection.otherName];
                conn.otherName = null;

                if(conn != null) {
                    sendTo(conn, {
                        Type: "leave"
                    });
                }
            }
        }
    });


});

function sendTo(connection, message) {
    connection.send(JSON.stringify(message));
}

*/