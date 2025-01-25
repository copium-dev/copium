import type { PageServerLoad } from './$types';
import { BACKEND_URL } from '$env/static/private';
import { redirect } from '@sveltejs/kit';
import type { Actions } from './$types';

// load function 
export const load: PageServerLoad = async ({ fetch }) => {
    const response = await fetch(`${BACKEND_URL}/user/dashboard`, {
        credentials: 'include'  // every protected route needs to include credentials
    });
    
    if (!response.ok) {
        throw redirect(303, `${BACKEND_URL}/auth/google`);
    }

    const applications = await response.json();
    
    return {
        applications
    };
};

export const actions = {
    add: async ({ request, fetch }) => {
        const formData = await request.formData();
        const data = {
            role: formData.get('role'),
            company: formData.get('company'),
            location: formData.get('location'),
            appliedDate: new Date(formData.get('appliedDate') as string),
            link: formData.get('link'),
            status: 'Applied'
        }

        const response = await fetch(`${BACKEND_URL}/user/addApplication`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify(data)
        });

        if (!response.ok) {
            return {
                type: 'error',
                message: 'Failed to add application'
            };
        }

        return {
            type: 'success',
            message: 'Application added successfully'
        };
    },
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
    editstatus: async ({ request, fetch }) => {
        const formData = await request.formData();
        const body = {
            id: formData.get('id'),
            status: formData.get('status')
        }

        const response = await fetch(`${BACKEND_URL}/user/editStatus`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify(body)
        });

        if (!response.ok) {
            return {
                type: 'error',
                message: 'Failed to update application status'
            };
        }
        
        return {
            type: 'success',
            message: 'Application status updated successfully'
        };
    },
} satisfies Actions;