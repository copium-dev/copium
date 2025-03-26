import type { RequestHandler } from './$types';
import { BACKEND_URL } from '$env/static/private';

export const POST: RequestHandler = async ({ request, fetch, locals }) => {
    const formData = await request.formData();
    const body = {
        id: formData.get('id'),
        operationID: formData.get('operationID')
    }

    const response = await fetch(`${BACKEND_URL}/user/revertStatus`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${locals.authToken}`
        },
        body: JSON.stringify(body)
    });

    if (!response.ok) {
        return new Response(JSON.stringify({
            type: 'failure',
        }), {
            headers: { 'Content-Type': 'application/json' }
        })
    }

    const newStatus = await response.json();

    return new Response(JSON.stringify({
        type: 'success',
        data: newStatus
    }), {
        headers: { 'Content-Type': 'application/json' }
    })
}