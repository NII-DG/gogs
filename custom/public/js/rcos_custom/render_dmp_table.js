// CreateTableFromJSON is RCOS specific code.
// This generates a table to check the information of the DMP
// created by the user and displays it on the screen.
function CreateTableFromJSON(tbodyEle, dmp) {
    let rowNum = getRowNum(dmp) //行数
    let colNum = getColNum(dmp) //列数

    // create reseted
    let dmpData2DArr = createDmp2dArr(rowNum, colNum)
    let dmptable = fillDmp2dArr(dmpData2DArr, dmp, 0, 0)
    let colspanTable = fillColspan2DArr(dmptable, rowNum, colNum)
    let rowspanTable = fillRowspan2DArr(dmptable, colspanTable, rowNum, colNum)

    for (let i =0; i < rowNum; i++)  {
        let tr = document.createElement("tr");
        for (let j = 0; j < colNum; j++) {
            let dmpUnitVal = dmptable[i][j]
            if (dmpUnitVal != ""){
                let td = document.createElement("td");
                td.textContent = dmpUnitVal
                let colspanNum = colspanTable[i][j]
                if (colspanNum >1) {
                    td.colSpan = colspanNum
                }
                let rowspanNum = rowspanTable[i][j]
                if (rowspanNum >1) {
                    td.rowSpan = rowspanNum
                }
                td.classList.add("dmp_unit_data")
                tr.appendChild(td)
            }

        }
        tbodyEle.appendChild(tr)
    }
    return colNum
}

$(document).ready(function () {
    let tbodyEle = document.getElementById("dmp");
    let items = JSON.parse($('#items').val());
    (document.getElementById("title")).innerHTML = "dmp.json (" + items.schema + ")";
    let colNum = CreateTableFromJSON(tbodyEle, items)
    let dmpHeaderRight = document.getElementById("dmp_header_right");
    dmpHeaderRight.colSpan = colNum-1;
});

function IsArrayObject(data)  {
    return Array.isArray(data)
}

function IsObjectInArrayObject(data)  {
    if (IsArrayObject(data)){
        if (data.length>0) {
            if (IsObject(data[0])){
                return true
            }else{
                return false
            }
        }else{
            return false
        }
    }else{
        return false
    }
}

function IsObject(data)  {
    return (typeof data === "object")
}

function IsStringOrBoolOrNumber(data) {
    type = typeof data
    if (type === 'string') {
        return true
    } else if (type === 'number') {
        return true
    } else if (type === 'boolean') {
        return true
    } else {
        return false
    }
}

function createDmp2dArr(rowNum, colNum) {
    // create dmp2DArr 2d array
    var dmp2DArr = [] //DMP data 2D array
    for (let i =0; i < rowNum; i++)  {
        let dmpArrChild = []
        for (let j = 0; j < colNum; j++) {
            dmpArrChild.push("")
        }
        dmp2DArr.push(dmpArrChild)
    }
    return dmp2DArr
}

function fillDmp2dArr(dmp2DArr, dmp, raw_index, col_index) {
    // fill DMP data 2d Array
    for (const key in dmp) {
        const value = dmp[key]
        const record_num =  getLeafNum(value)

        endPoint = raw_index
        if (IsObjectInArrayObject(value)) {
            endPoint = endPoint + record_num
        }else {
            endPoint = endPoint + 1
        }
        for (i =raw_index; i< endPoint; i++) {
            dmp2DArr[i][col_index] = key
            if (record_num == 1) {
                dmp2DArr[i][col_index+1] = String(value)
            }else{
                if (IsArrayObject(value)) {
                    // is array
                    let arr_start_index = i
                    for (let nested_i = 0; nested_i < value.length; ++nested_i) {
                        const arr_data = value[nested_i]
                        if (IsObject(arr_data)) {
                            const arr_data_record_num =  getLeafNum(arr_data)
                            dmp2DArr[arr_start_index][col_index+1] = String(nested_i)
                            next_raw_index = arr_start_index
                            next_col_index = col_index+2
                            dmp2DArr = fillDmp2dArr(dmp2DArr, arr_data, next_raw_index, next_col_index)
                            arr_start_index = arr_start_index + arr_data_record_num
                        } else {
                            dmp2DArr[arr_start_index][col_index+1] = String(nested_i)
                            dmp2DArr[arr_start_index][col_index+2] = arr_data
                            arr_start_index = arr_start_index + 1
                        }
                    }

                } else {
                    next_raw_index = raw_index
                    next_col_index = col_index+1
                    dmp2DArr = fillDmp2dArr(dmp2DArr, value, next_raw_index, next_col_index)
                }

            }
        }
        raw_index = raw_index + record_num
    }
    return dmp2DArr
}

function createRowspanOrColspan2DArr(rowNum, colNum) {
        // create resed each 2d array
        var rowspanOrColspan2DArr = [] //rowspan info 2D array
        for (let i =0; i < rowNum; i++)  {
            let rowspanArrChild = []
            for (let j = 0; j < colNum; j++) {
                rowspanArrChild.push(0)
            }
            rowspanOrColspan2DArr.push(rowspanArrChild)
        }
    return rowspanOrColspan2DArr
}

function isEmptyInNextCol(arr, index){
    const arrLength = arr.length
    const nextIndex = index+1
    if (arrLength==nextIndex) {
        return false
    }else {
        const nextVal = arr[nextIndex]
        if (nextVal==""){
            return true
        }else{
            return false
        }
    }

}

function fillColspan2DArr(dmpDataTable, rowNum, colNum) {
    let colspan2DArr =createRowspanOrColspan2DArr(rowNum, colNum)
    for (let i=0; i < rowNum; i++) {
        let rowDmpData = dmpDataTable[i]
        for (let j = 0; j < colNum; j++) {
            if (rowDmpData[j]==""){
                continue
            }else{
                if (isEmptyInNextCol(rowDmpData, j)){
                    let colspanNum = colNum - j
                    colspan2DArr[i][j] = colspanNum
                    break
                }else{
                    continue
                }
            }

        }
    }
    return colspan2DArr
}

function hasColspanBefore(colspanDataArry, targetIndex){
    for (let i=0; i < targetIndex; i++) {
        if (colspanDataArry[i]==0){
            continue
        }else{
            return true
        }
    }
    return false
}

function getRowspan(dmpDataTable, colspanTable, startRowIndex, colIndex) {
    let rowspanNum = 0
    for (let i=startRowIndex; i < dmpDataTable.length; i++) {
        let targetVal = dmpDataTable[i][colIndex]
        if (targetVal==""){
            //check colspanTable
            if (hasColspanBefore(colspanTable[i], colIndex)){
                return rowspanNum+1
            }
            rowspanNum = rowspanNum +1
        }else{
            break
        }
    }
    return rowspanNum +1
}

function fillRowspan2DArr(dmpDataTable, colspanTable, rowNum, colNum) {
    let rowspan2DArr =createRowspanOrColspan2DArr(rowNum, colNum)
    for (let i=0; i < rowNum; i++) {
        let rowDmpData = dmpDataTable[i]
        for (let j = 0; j < colNum; j++) {
            let arrVal = rowDmpData[j]
            if (arrVal==""){
                continue
            }else{
                rowspanNum = getRowspan(dmpDataTable, colspanTable, i+1, j)
                rowspan2DArr[i][j] = rowspanNum
            }
        }
    }
    return rowspan2DArr
}





function getLeafNum(data) {
    if (IsStringOrBoolOrNumber(data)) {
        return 1
    } else {
        return getRowNum(data)
    }
}

function getRowNum(data) {
    //get creating row num
    let row_num = 0
    for (const key in data) {
        let value = data[key]
        if (IsArrayObject(value)) {
            // is array
            for (let i = 0; i < value.length; ++i) {
                let arr_val = value[i]
                if  (IsObject(arr_val)) {
                    row_num = row_num + getRowNum(value[i])
                }else{
                    row_num = row_num + 1
                }
            }
        } else if (IsObject(value)) {
            // is json object
            row_num = row_num + getRowNum(value)
        } else {
            // is str, num, bool...
            row_num = row_num + 1
        }
    }
    return row_num
}

function getColNum(data) {
    //get creating col num
    return getNestedNum(data, 0) + 2 //列数
}

function getNestedNum(data, pre_nested_num) {
    // get nested num in dmp json
    let nested_num = pre_nested_num

    for (key in data) {
        //console.log("key : " + key)
        let value = data[key]
        if (IsArrayObject(value)) {
            // is array
            for (let i = 0; i < value.length; ++i) {
                let arr_val = value[i]
                if (IsObject(arr_val)){
                    arr_tmp = getNestedNum(arr_val, pre_nested_num+1) + 1
                    if (arr_tmp > nested_num) {
                        nested_num = arr_tmp
                    }
                }
            }
        } else if (IsObject(value)) {
            // is json object
            tmp = getNestedNum(value, pre_nested_num+1)
            if (tmp > nested_num) {
                nested_num = tmp
            }
        } else {
            // is str, num, bool...
        }
    }
    return nested_num
}