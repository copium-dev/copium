import type { PageServerLoad } from './$types';
import { BACKEND_URL } from '$env/static/private';
import { redirect } from '@sveltejs/kit';
import type { Actions } from './$types';

// load function 
export const load: PageServerLoad = async ({ fetch }) => {
    const response = await fetch(`${BACKEND_URL}/user/profile`, {
        credentials: 'include'  // every protected route needs to include credentials
    });
    
    if (!response.ok) {
        throw redirect(303, `${BACKEND_URL}/auth/google`);
    }

    const data = await response.json();
    
    return {
        email: data.email,
        applicationsCount: data.applicationsCount,
    };
};

export const actions = {
    delete: async ({ fetch }) => {
        const response = await fetch(`${BACKEND_URL}/user/deleteUser`, {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({}),
        });

        if (!response.ok) {
            throw new Error('Failed to delete user');
        }
    },
} satisfies Actions;