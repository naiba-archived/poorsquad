$('.ui.checkbox').checkbox();
$('.ui.dropdown').dropdown();
$('.ui.avatar.dropdown').dropdown({
    action: 'nothing',
    on: 'hover',
});

const confirmBtn = $('.mini.confirm.modal .positive.button')
function showConfirm(title, content, callFn, extData) {
    const modal = $('.mini.confirm.modal')
    modal.children('.header').text(title)
    modal.children('.content').text(content)
    if (confirmBtn.hasClass('loading')) {
        return false
    }
    modal.modal({
        closable: true,
        onApprove: function () {
            confirmBtn.toggleClass('loading')
            callFn(extData)
            return false
        }
    }).modal('show')
}

function showFormModal(modelSelector, formID, URL, getData) {
    $(modelSelector).modal({
        closable: true,
        onApprove: function () {
            let success = false
            const btn = $(modelSelector + ' .positive.button')
            const form = $(modelSelector + ' form')
            if (btn.hasClass('loading')) {
                return success
            }
            form.children('.message').remove()
            btn.toggleClass('loading')
            const data = getData ? getData() : $(formID).serializeArray().reduce(function (obj, item) {
                obj[item.name] = (item.name.endsWith('_id') || item.name === 'id' || item.name === 'permission') ? parseInt(item.value) : item.value;
                return obj;
            }, {});
            $.post(URL, JSON.stringify(data)).done(function (resp) {
                if (resp.code == 200) {
                    if (resp.message) {
                        alert(resp.message)
                    }
                    window.location.reload()
                } else {
                    form.append(`<div class="ui negative message"><div class="header">操作失败</div><p>` + resp.message + `</p></div>`)
                }
            }).fail(function (err) {
                form.append(`<div class="ui negative message"><div class="header">网络错误</div><p>` + err.responseText + `</p></div>`)
            }).always(function () {
                btn.toggleClass('loading')
            });
            return success
        }
    }).modal('show')
}

function deleteRequest(api) {
    $.ajax({
        url: api,
        type: 'DELETE',
    }).done(resp => {
        if (resp.code == 200) {
            if (resp.message) {
                alert(resp.message)
            } else {
                alert('移除成功')
            }
            window.location.reload()
        } else {
            alert('移除失败 ' + resp.code + '：' + resp.message)
            confirmBtn.toggleClass('loading')
        }
    }).fail(err => {
        alert('网络错误：' + err.responseText)
    });
}

function addTeam() {
    showFormModal('.tiny.team.modal', '#teamForm', '/api/team');
}

function bindRepository(id, repos) {
    $('#bindRepositoryForm input[name=id]').val(id)
    $('#bindRepositoryForm .checkbox').checkbox('uncheck')
    if (repos) {
        for (let i = 0; i < repos.length; i++) {
            $('#bindRepositoryForm .id-' + repos[i]).checkbox('check')
        }
    }
    getData = function () {
        return $('#bindRepositoryForm').serializeArray().reduce(function (obj, item) {
            if (!obj['repositories']) {
                obj['repositories'] = []
            }
            if (item.value === 'on') {
                obj['repositories'].push(parseInt(item.name))
            } else {
                obj[item.name] = (item.name.endsWith('_id') || item.name === 'id') ? parseInt(item.value) : item.value;
            }
            return obj;
        }, {})
    }
    showFormModal('.tiny.bind-repository.modal', '#bindRepositoryForm', '/api/team/repositories', getData)
}

function addAccount() {
    showFormModal('.tiny.account.modal', '#accountForm', '/api/account');
}

function addCompany() {
    showFormModal('.tiny.company.modal', '#companyForm', '/api/company');
}

function addOrEditRepository(isEdit, repo) {
    const modal = $('.tiny.repository.add.modal')
    console.log(repo, $('#repositoryForm input[name=name]'))
    modal.children('.header').text((isEdit ? '修改' : '创建') + '仓库')
    modal.find('.positive.button').html((isEdit ? '修改' : '创建') + '<i class="add icon"></i>')
    modal.find('input[name=id]').val(isEdit ? repo.ID : null)
    modal.find('input[name=name]').val(isEdit ? repo.Name : null)
    if (isEdit) {
        modal.find('.private.dropdown').dropdown('set selected', repo.Private ? 'on' : 'off')
        modal.find('.account.dropdown').dropdown('set selected', repo.AccountID)
        modal.find('.account.dropdown').addClass('disabled')
    } else {
        modal.find('.account.dropdown').removeClass('disabled')
        modal.find('.private.dropdown').dropdown('restore defaults')
        modal.find('.account.dropdown').dropdown('restore defaults')
    }
    showFormModal('.tiny.repository.add.modal', '#repositoryForm', '/api/repository');
}

function addEmployee(type, id) {
    $('#employeeForm .dropdown').parent().css('display', 'block')
    $('#employeeForm .dropdown .item:nth-child(1)').css('display', 'block')
    $('#employeeForm .dropdown .item:nth-child(3)').css('display', 'block')
    $('#employeeForm input[name=id]').val(id)
    $('#employeeForm input[name=type]').val(type)
    $('#employeeForm .dropdown').dropdown('restore default value')
    if (type === 'repository') {
        $('#employeeForm .dropdown').parent().css('display', 'none')
    } else if (type === 'team') {
        $('#employeeForm .dropdown .item:nth-child(3)').css('display', 'none')
    } else {
        $('#employeeForm .dropdown .item:nth-child(1)').css('display', 'none')
    }
    showFormModal('.tiny.employee.modal', '#employeeForm', '/api/employee');
}

function removeTeam(id) {
    deleteRequest('/api/team/' + id)
}

function removeRepository(id, name) {
    const modal = $('.repository.delete.modal')
    const form = $('.repository.delete.modal form')
    const btn = $('.repository.delete.modal .positive.button')
    const nameEl = $('#deleteRepositoryForm input')
    modal.children('.header').text("确认删除仓库「" + name + "」？")
    if (btn.hasClass('loading')) {
        return false
    }
    modal.modal({
        closable: true,
        onApprove: function () {
            form.children('.message').remove()
            if (nameEl.val() !== name) {
                form.append(`<div class="ui negative message"><div class="header">操作失败</div><p>仓库名称不匹配</p></div>`)
                return false
            }
            btn.toggleClass('loading')
            $.ajax({
                url: '/api/repository/' + id + '/' + name,
                type: 'DELETE',
            }).done(resp => {
                if (resp.code == 200) {
                    if (resp.message) {
                        alert(resp.message)
                    } else {
                        alert('移除成功')
                    }
                    window.location.reload()
                } else {
                    alert('移除失败 ' + resp.code + '：' + resp.message)
                    btn.toggleClass('loading')
                }
            }).fail(err => {
                form.append(`<div class="ui negative message"><div class="header">网络错误</div><p>` + err.responseText + `</p></div>`)
            });
            return false
        }
    }).modal('show')
}

function removeEmployee(data) {
    deleteRequest('/api/employee/' + data.type + '/' + data.id + '/' + data.userID)
}

function logout(id) {
    $.post('/api/logout', JSON.stringify({ id: id })).done(function (resp) {
        if (resp.code == 200) {
            alert('注销成功')
            window.location.href = '/login'
        } else {
            alert('注销失败 ' + resp.code + '：' + resp.message)
        }
    }).fail(function (err) {
        alert('网络错误：' + err.responseText)
    })
}
