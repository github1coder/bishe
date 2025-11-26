import request from '@/api/request.js'
import { useApiStore } from '@/stores/api.js'
import { useUserStore } from '@/stores/user'

const baseUrl = '/domain'

/**
 * 创建数据域
 * @param {string} domainName - 数据域名称
 * @returns {Promise}
 */
export function createDomain(domainName) {
    const apiStore = useApiStore()
    const userStore = useUserStore()
    return request({
        url: baseUrl + '/createDomain',
        method: 'post',
        data: {
            apiUrl: {
                ipfsServiceUrl: apiStore.ipfsServiceUrl,
                contractName: apiStore.contractName,
                chainServiceUrl: apiStore.chainServiceUrl,
            },
            domainName: domainName,
            orgId: userStore.orgId,
            role: userStore.role,
        },
    })
}

/**
 * 更新数据域元数据
 * @param {string} domainName - 数据域名称
 * @param {string} domainMembers - 数据域成员
 * @param {string} domainPolicy - 数据域策略
 * @returns {Promise}
 */
export function updateDomainMetadata(domainName, domainMembers, domainPolicy) {
    const apiStore = useApiStore()
    const userStore = useUserStore()
    return request({
        url: baseUrl + '/updateDomainMetadata',
        method: 'post',
        data: {
            apiUrl: {
                ipfsServiceUrl: apiStore.ipfsServiceUrl,
                contractName: apiStore.contractName,
                chainServiceUrl: apiStore.chainServiceUrl,
            },
            domainName: domainName,
            domainMembers: domainMembers,
            domainPolicy: domainPolicy,
            orgId: userStore.orgId,
            role: userStore.role,
        },
    })
}

/**
 * 查询当前orgId下的数据域
 * @returns {Promise}
 */
export function queryMyDomains() {
    const apiStore = useApiStore()
    const userStore = useUserStore()
    return request({
        url: baseUrl + '/queryMyDomains',
        method: 'post',
        data: {
            apiUrl: {
                ipfsServiceUrl: apiStore.ipfsServiceUrl,
                contractName: apiStore.contractName,
                chainServiceUrl: apiStore.chainServiceUrl,
            },
            orgId: userStore.orgId,
            role: userStore.role,
        },
    })
}

/**
 * 查询当前orgId下管理的数据域
 * @returns {Promise}
 */
export function queryMyManagedDomains() {
    const apiStore = useApiStore()
    const userStore = useUserStore()
    return request({
        url: baseUrl + '/queryMyManagedDomains',
        method: 'post',
        data: {
            apiUrl: {
                ipfsServiceUrl: apiStore.ipfsServiceUrl,
                contractName: apiStore.contractName,
                chainServiceUrl: apiStore.chainServiceUrl,
            },
            orgId: userStore.orgId,
            role: userStore.role,
        },
    })
}

/**
 * 根据数据域名查询数据域详细信息
 * @param {string} domainName - 数据域名称
 * @returns {Promise}
 */
export function queryDomainInfo(domainName) {
    const apiStore = useApiStore()
    const userStore = useUserStore()
    return request({
        url: baseUrl + '/queryDomainInfo',
        method: 'post',
        data: {
            apiUrl: {
                ipfsServiceUrl: apiStore.ipfsServiceUrl,
                contractName: apiStore.contractName,
                chainServiceUrl: apiStore.chainServiceUrl,
            },
            domainName: domainName,
            orgId: userStore.orgId,
            role: userStore.role,
        },
    })
}

