const express = require("express");
const ldap = require("ldapjs");

const LDAP_URL = 'ldap://localhost:389';
const BASE_DN = 'dc=mokapi,dc=io';
const PORT = 3000;

const client = ldap.createClient({ url: LDAP_URL });

function authenticate(username, password, callback) {
    const dn = `uid=${username},${BASE_DN}`;
    client.bind(dn, password, (err) => {
        if (err) return callback(null, false);
        callback(null, true);
    });
}

function checkGroup(username, groupName, callback) {
    const searchOptions = {
        filter: `(memberOf=cn=${groupName},ou=groups,${BASE_DN})`,
        scope: 'sub'
    };

    client.search(`cn=${username},${BASE_DN}`, searchOptions, (err, res) => {
        let found = false;
        res.on('searchEntry', () => found = true);
        res.on('end', () => callback(null, found));
    });
}

const app = express();
app.use(express.json());

app.post('/login', (req, res) => {
    const { username, password } = req.body;
    authenticate(username, password, (err, success) => {
        if (!success) return res.status(401).send('Invalid credentials');
        checkGroup(username, 'Admins', (err, isAdmin) => {
            if (isAdmin) {
                res.send({ message: 'Welcome, Admin!' });
            } else {
                res.status(403).send('Access denied');
            }
        });
    });
});

app.listen(PORT, () => console.log(`Server running on port ${PORT}`));