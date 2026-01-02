export async function driveHttp() {
  await fetch('http://localhost/v2/pet/10', { headers: { Accept: 'application/json', api_key: 'demo' }})
  await fetch('http://localhost/v2/pet', {
    method: 'POST',
    body: JSON.stringify({ name: 'Milo', photoUrls: [] }),
    headers: {
      Authorization: 'demo',
      'Content-Type': 'application/json'
    }
  });

  await fetch('http://localhost/v2/user/bmiller');

  // invalid request without password
  await fetch('http://localhost/v2/user/login?username=ajohnson');
  // valid
  await fetch('http://localhost/v2/user/login?username=ajohnson&password=anothersecretpassword456');
  // invalid password
  await fetch('http://localhost/v2/user/login?username=bmiller&password=12345');
  await fetch('http://localhost/v2/user/login?username=bmiller&password=mysecretpassword123');
  // request with script error
  await fetch('http://localhost/v2/user/login?username=bmiller&password=mysecretpassword123');
  // script error
  await fetch('http://localhost/v2/user/logout');

  // should only return pets with the given status.
  await fetch('http://localhost/v2/pet/findByStatus?status=available,pending', {
    headers: {
      Accept: 'application/xml',
      Authorization: 'demo'
    }
  });
  
  await fetch('http://localhost/v2/store/order/1')
  await fetch('http://localhost/v2/store/order', {
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      id: 2,
      petId: 14921,
      quantity: 2,
      status: 'placed'
    })
  })
  await fetch('http://localhost/v2/store/order/2', {
    method: 'DELETE',
  })
}