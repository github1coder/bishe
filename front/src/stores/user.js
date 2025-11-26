import { ref } from 'vue';
import { defineStore } from 'pinia';

export const useUserStore = defineStore('user', () => {
    const userId = ref('user1111');
    const orgId = ref('hospitalA');
    const role = ref('doctor');

    function setUserId(id) {
        userId.value = id;
    }

    function setOrgId(id) {
        orgId.value = id;
    }

    function setRole(roleName) {
        role.value = roleName;
    }

    function setOrgRole(id, roleName) {
        orgId.value = id;
        role.value = roleName;
    }

    return { userId, orgId, role, setUserId, setOrgId, setRole, setOrgRole };
});
