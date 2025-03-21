// THIS IS EXTREMELY EXTREMELY TEMPORARY
//  This is a currently working solution to the cross-domain cookie issue. Cookies
//  are preferred because they are more secure and make the auth flow a bit simpler.
//  Technically, Cloud Run does support custom domains, but it's in preview mode and
//  is not recommended for production yet. So, until that's available, we have to
//  use the traditional method of setting tokens in localStorage.
//  We do have a hybrid solution though where we use a cookie to store the token
//  because all requests to the backend are made from +page.server.ts which can access
//  cookies. so, no fetch requests are changed from frontend, only SvelteKit server-side
//  functions have been modified to use the cookie set here
import type { RequestHandler } from '@sveltejs/kit';

export const GET: RequestHandler = ({ url }) => {
    const token = url.searchParams.get('token');
    // token exists -- redirect to dashboard with token in a cookie
    // the browser sets the cookie to be gotten by hooks which are then used in +page.server.ts
    if (token) {
        console.log('authed', token)
        return new Response(null, {
            status: 302,
            headers: {
                'Location': '/dashboard',
                'Set-Cookie': `authToken=${token}; Path=/; Max-Age=${30*24*60*60}; SameSite=Lax; Secure`
            }
        });
    }
    
    // no token -- redirect home. set cookie to nothing just in case
    return new Response(null, {
        status: 302,
        headers: {
            'Location': '/',
            'Set-Cookie': `authToken=; Path=/; Max-Age=0; SameSite=Lax; Secure`
        }
    });
};