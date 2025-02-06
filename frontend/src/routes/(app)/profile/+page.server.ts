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

    const applications = await response.json();
    
    return {
        email: applications.email,
    };
};

export const actions = {
    delete: async ({ request, fetch }) => {
        const formData = await request.formData();
        const body = formData.get('id');

        const response = await fetch(`${BACKEND_URL}/user/deleteApplication`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify({ id: body })
        });

        if (!response.ok) {
            return {
                type: 'error',
                message: 'Failed to delete application'
            };
        }
        
        return {
            type: 'success',
            message: 'Application deleted successfully'
        };
    },
} satisfies Actions;