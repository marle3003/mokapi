const express = require("express");
const ldap = require("ldapjs");

const app = express();
app.use(express.json());

const LDAP_URL = 'ldap://localhost:389';
const BASE_DN = 'dc=mokapi,dc=io';

function authenticate(username, password, callback) {
    const client = ldap.createClient({ url: LDAP_URL });
    const userDn = `uid=${username},${BASE_DN}`;

    client.bind(userDn, password, (err) => {
        client.unbind();
        if (err) {
            return callback(false);
        }
        callback(true);
    });
}

app.post("/login", (req, res) => {
    const { username, password } = req.body;
    authenticate(username, password, (success) => {
        if (success) {
            res.json({ message: "Authentication successful" });
        } else {
            res.status(401).json({ message: "Invalid credentials" });
        }
    });
});

const PORT = 3000;
app.listen(PORT, () => {
    console.log(`Server running on http://localhost:${PORT}`);
});