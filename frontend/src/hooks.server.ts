import type { Handle } from '@sveltejs/kit';

// since this is a hook it runs for every request
// we need this to allow any server-side action/load function to access
// via locals.authToken
export const handle: Handle = async ({ event, resolve }) => {
    // get token from cookie (set in auth-complete)
    const authToken = event.cookies.get('authToken');
    
    if (authToken) {
        event.locals.authToken = authToken;
    }
    
    return resolve(event);
};