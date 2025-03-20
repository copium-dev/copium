import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
    // get token from cookie (set in auth-complete)
    const authToken = event.cookies.get('authToken');
    
    if (authToken) {
        event.locals.authToken = authToken;
    }
    
    return resolve(event);
};