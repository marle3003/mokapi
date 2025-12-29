export async function driveHttp() {
  await fetch('http://localhost/v2/pet/10', { headers: { api_key: 'demo' }})
  await fetch('http://localhost/v2/pet', {
    method: 'POST',
    body: JSON.stringify({ name: 'Milo', photoUrls: [] }),
    headers: {
      Authorization: 'demo',
      'Content-Type': 'application/json'
    }
  });

  await fetch('http://localhost/v2/user/bmiller');
}