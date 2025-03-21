import type { RequestHandler } from '@sveltejs/kit';

// remove cookie and redirect home 
export const GET: RequestHandler = () => {
    return new Response(null, {
        status: 302,
        headers: {
            'Location': '/',
            'Set-Cookie': `authToken=; Path=/; Max-Age=0; SameSite=Lax; Secure`
        }
    });
};