import fs from 'fs/promises'

const output = '../../public/demo'

export async function collectBruno() {
    const url = 'http://localhost:8080/api/services/http/Swagger%20Petstore/bruno.yaml?host=localhost%3A8080&folderArrangement=tags&itemName=summary';
    const res = await fetch(url)

    await fs.writeFile(output + '/bruno.yaml', await res.text());
}