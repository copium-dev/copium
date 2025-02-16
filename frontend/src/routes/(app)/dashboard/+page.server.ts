import type { PageServerLoad } from './$types';
import { BACKEND_URL } from '$env/static/private';
import { redirect } from '@sveltejs/kit';
import type { Actions } from './$types';

// load function 
export const load: PageServerLoad = async ({ fetch, url }) => {
    const page = url.searchParams.get('page');
    const query = url.searchParams.get('q');
    const company = url.searchParams.get('company');
    const status = url.searchParams.get('status');
    const role = url.searchParams.get('role');
    const location = url.searchParams.get('location');
    const startDate = url.searchParams.get('startDate');
    const endDate = url.searchParams.get('endDate');

    const params = new URLSearchParams();
    if (page) params.set('page', page);
    if (query) params.set('q', query);
    if (company) params.set('company', company);
    if (status) params.set('status', status);
    if (role) params.set('role', role);
    if (location) params.set('location', location);
    if (startDate) params.set('startDate', startDate);
    if (endDate) params.set('endDate', endDate);

    const dashboardURL = `${BACKEND_URL}/user/dashboard?${params.toString()}`;

    const response = await fetch(dashboardURL, {
        credentials: 'include'  // every protected route needs to include credentials
    });
    
    if (!response.ok) {
        throw redirect(303, `${BACKEND_URL}/auth/google`);
    }

    const data = await response.json();

    const applications = data.applications || [];
    const currentPage = data.currentPage || parseInt(page || '1');
    const totalPages = data.totalPages || 1;
    const clientParams = params.toString();

    {console.log("applications:" + applications)}
    {console.log("currentPage:" + currentPage)}
    {console.log("totalPages:" + totalPages)}
    {console.log("clientParams:" + clientParams)}
    
    return {
        applications,
        currentPage,
        totalPages,
        clientParams,
    };
};

export const actions = {
    add: async ({ request, fetch }) => {
        const formData = await request.formData();
        const data = {
            role: formData.get('role'),
            company: formData.get('company'),
            location: formData.get('location'),
            appliedDate: Date.parse(formData.get('appliedDate') as string),
            link: formData.get('link'),
            status: 'Applied'
        }

        {console.log(data)}

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
    editapplication: async({ request, fetch }) => {
        const formData = await request.formData();
        const body = {
            id: formData.get('id'),
            role: formData.get('role'),
            company: formData.get('company'),
            location: formData.get('location'),
            appliedDate: Date.parse(formData.get('appliedDate') as string),
            link: formData.get('link'),
        }

        const response = await fetch(`${BACKEND_URL}/user/editApplication`, {
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
                message: 'Failed to update application'
            };
        }

        throw redirect(303, '/dashboard');
    }
} satisfies Actions;