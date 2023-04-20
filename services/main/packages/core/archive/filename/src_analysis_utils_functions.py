# Library for common functions used by various scripts
from pathlib import Path
import re
import sys




## -------------------------------------------------------------------------------
## PURPOSE:
##   Returns True if at least one character in string is a digit
## ARGS:
##   input_string - the string to check
## RETURNS:
##   True - if it has at least one digit
##   False -if it does not
## -------------------------------------------------------------------------------

def has_digits(input_string):
    return bool(re.search(r'\d', input_string))



## -------------------------------------------------------------------------------
## PURPOSE:
##   Rreturns package name and version based on the package name.
##   For example:
##     WebTest-2.0.35.tar.gz  --> ("webtest", "2.0.35")
##     urllib3-1.26.5.tar.gz -> ("urllib3", "1.26.5")
## ARGS:
##   filename - archive name
## RETURNS:
##   ("<name-substring>", "<version>")
## -------------------------------------------------------------------------------

def getPkgNameVersion (filename):
    odd_str_list = ['ubuntu', '+dfsg', '.dfsg', '+deb', '+nmu', '+git',  '~', '+', '.stx', '.tar.gz', '.tar.bz2', '.tar.xz', '.zip']

    name_lowercase = filename.lower()
    ## remove noise in name

    ## case 1:
    ## for the special case were an extra _#_ is used, for example:
    ##  '_2_' in file_2_5.93  or
    #   '_1_' in libjpeg-turbo_1_2.0.tar.gz
    noise_str_list = re.findall("_[0-9]*_", name_lowercase)  ## e.g., = ['_2_'] in file_2_5.93
    for item in noise_str_list:
        ## replace noisy string with '_'
        name_lowercase = name_lowercase.replace (item, "_")

    ## Parse the file name
    dash_delimiter_used       = False
    underscore_delimiter_used = False
    ## Next decide if the name subcomponents are deliminated by '-' or '_' (the two most common)
    ## e.g.,  ceph-manager-1.0-26.rpm  vs. curl_7.74.0-1.3_deb11u3.tar.gz
    if name_lowercase.find ("-") == -1 and name_lowercase.find ("_") == -1:
          ## Neither '-", '_' deliminator found
          ## assume delimiter '.'
          filename_list = name_lowercase.split('.')
    if name_lowercase.find ("_") == -1 or \
       (name_lowercase.find ("-") > -1 and name_lowercase.find ("-") < name_lowercase.find ("_")):
      ## dash ('-') used as primary delimiter
      dash_delimiter_used = True
      filename_list = name_lowercase.split('-')

    # elif name_lowercase.find ("-") == -1 and name_lowercase.find ("_") > -1:
    #   filename_list = name_lowercase.split('_')
    #   dash_delimiter_used = false
    elif name_lowercase.find ("-") == -1 or \
         (name_lowercase.find ("_") > -1 and name_lowercase.find ("_") < name_lowercase.find ("-")):
      ## Underscore '_' was used as delimeter
      underscore_delimiter_used = True
      filename_list = name_lowercase.split('_')

    #   if len (filename_list) > 1:
    #     ## This is specific to the '_' delimeter where subparts > 1
    #     ## for the special case were an extra _#_ is used (e.g. _1_ in file_1_5.93)
    #     ## we remove it.
    #     count = 0
    #     for item in filename_list:
    #         if item.isdigit():
    #             ## e.g., file_1_5.39-3.tar.gz -> ['file', '1', '5.39-3.tar.gz']
    #             del filename_list[count]   # delete the item (e.g, '1')
    #             ## file_1_5.39-3.tar.gz -> ['file','5.39-3.tar.gz']
    #             break  # quit after finding the first one (e.g. _1_ in file_1_5.93)
    #         count += 1
    else:  ## I have no idea
        return ('', '')

    # if "libjpeg-turbo" in filename or "file" in filename or "adwaita-icon-theme" in filename:
    #     print (filename_list)

    num_subparts = len (filename_list)
    display_name = ''
    count = 0
    #### print ("---", filename, filename_list)
    for subpart in filename_list:
        ## for each subpart
        if len(subpart) == 0:
            ## some subparts can be '' because they have '--' in name - e.g. becke-ch--regex--s0-0-v1--base--pl--lib-1.4.0.zip,
            continue
        else:
            count += 1

        if count == 1: ## for FIRST subpart
            ## next check when only one sub part - e.g., no version present
            if num_subparts == 1: ## e.g.,  myprogram.zip, utils.tar.gz app.1.2.zip
                print ("only one subpart:", subpart, filename_list, file=sys.stderr )
                ## 'app.zip' --> ['.zip']  utils.tar.gz --> ['.tar', '.gz']   velero.1.2.zip --> ['.1', '.2', '.zip']
                extensions_list = Path(subpart).suffixes
                if len(extensions_list) == 0: ##  myapp (rare if at all)
                    display_name += subpart + ':'
                elif len(extensions_list) == 1:  ## 'app.zip' --> ['.zip']
                    index = subpart.find (extensions_list[0])
                    display_name += subpart[:index] + ':'
                elif len(extensions_list) >= 2 and '.tar' in extensions_list: ## utils.tar.gz --> ['.tar', '.gz'], app.1.2.tar.gz --> ['.1', '.2', '.tar', '.gz']
                    ##index = subpart.find (extensions_list[0] + extensions_list[1])
                    all_extensions = ''.join(extensions_list) ## --> '.1.2.tar.gz'
                    ext1 = extensions_list.pop()
                    ext2 = extensions_list.pop()
                    remaing_extensions = ''.join(extensions_list) ## --> '.1.2'
                    print ("extx", ext1, ext2, extensions_list, display_name, file=sys.stderr)
                    index = subpart.find (all_extensions) ## index for '.1.2.tar.gz'
                    display_name +=  subpart[:index] + ':' + remaing_extensions [1:] ## [1:] removes '.' from '.1.2' --> --> '1.2'
                    print (f"here: {display_name} |{ext2}{ext1}| {subpart[:index]}", file=sys.stderr)
                else: ## > 2 two where second is not '.tar'   e.g.,  app.1.2.3.gz --> ['.1', '.2', '.3', '.gz']  --> ('app','1.2.3')
                    extensions_list.pop() ## pop '.gz'
                    remaing_extensions = ''.join(extensions_list) ## --> '.1.2.3'
                    index = subpart.find (remaing_extensions) ## index for '.1.2.3'
                    display_name += subpart[:index] + ':' + remaing_extensions [1:] ## [1:] removes '.' from '.1.2.3' --> '1.2.3'
                found = True
                break
            else:
                ## two or more subparts - assume first subpart is part of component name
                display_name += subpart
            continue


        ## Work on the version part of the file name
        if '.' in subpart or subpart[0].isdigit():  ## if '.' or first char is digit
            found_verion = False
            ## remove odd strings if any found.
            for odd_str in odd_str_list:
                index = subpart.find (odd_str)
                if index != -1:
                    ## slice off odd_str
                    subpart_slice =  subpart[:index]
                    subpart = subpart_slice

            ## next check if this is the last subpart
            if count == num_subparts and not has_digits(subpart):  ## if last subpart and no digits
                ## no version found in last subpart (i.e., no digits)
                display_name += "-" + subpart + ':'
            else: ## not last but found version OR last piece with version
                display_name += ':' + subpart
            found_verion = True

            if found_verion == False:
                display_name += ':' + subpart
            break
        else: ## '.' not found and no digits
            display_name += "-" + subpart ## continue adding subparts
    ## display name consists of component name + version seperated by a ':'. Obtain both
    name_version_list = display_name.split (':')

    if len (name_version_list) > 1:
        name = name_version_list[0]
        version = name_version_list[1]
    elif len (name_version_list) == 1:
        name = name_version_list[0]
        version = ""
    else:
        name = ''
        version = ''

    ## When '-' is the main delimeter but there is a '_' BEFORE version we need to clean it up
    ## For example:
    ##                                           name            version
    ##  init-system-helpers_1.60.tar.gz   --> ['init-system', 'helpers_1.60']
    ##    we want                         --> ['init-system-helpers', '1.60']
    if dash_delimiter_used and '_' in version:
        ## split up version using the '_' delimeter and adding first part back to the name
        version_parts = version.split ('_')
        name += '-' + version_parts[0]
        version = version_parts[1]

    return (name, version)




## -------------------------------------------------------------------------------
## PURPOSE:
##   Rreturns package name and version based on the package name.
##   For example:
##     WebTest-2.0.35.tar.gz  --> ("webtest", "2.0.35")
##     urllib3-1.26.5.tar.gz -> ("urllib3", "1.26.5")
## ARGS:
##   filename - archive name
## RETURNS:
##   ("<name-substring>", "<version>")
## -------------------------------------------------------------------------------

def getPkgNameVersion3 (filename):
    odd_str_list = ['ubuntu', '+dfsg', '.dfsg', '+deb', '+nmu', '+git',  '~', '+', '.stx', '.tar.gz', '.tar.bz2', '.tar.xz', '.zip']
    name_lowercase = filename.lower()

    ## First dediced if the name subcomponents are deliminated by '-' or '_' (the two most common)
    ## e.g.,  ceph-manager-1.0-26.rpm  vs. curl_7.74.0-1.3_deb11u3.tar.gz
    if name_lowercase.find ("-") == -1 and name_lowercase.find ("_") == -1:
          ## Neither deliminator found
          return ('', '')
    if name_lowercase.find ("_") == -1 or \
       (name_lowercase.find ("-") > -1 and name_lowercase.find ("-") < name_lowercase.find ("_")):
      ## dash ('-') used as primary delimiter
      dash_delimiter_used = True
      filename_list = name_lowercase.split('-')

    # elif name_lowercase.find ("-") == -1 and name_lowercase.find ("_") > -1:
    #   filename_list = name_lowercase.split('_')
    #   dash_delimiter_used = false
    elif name_lowercase.find ("-") == -1 or \
         (name_lowercase.find ("_") > -1 and name_lowercase.find ("_") < name_lowercase.find ("-")):
      ## Underscore '_' was used as delimeter
      dash_delimiter_used = False
      filename_list = name_lowercase.split('_')

      if len (filename_list) > 1:
        ## This is specific to the '_' delimeter where subparts > 1
        ## for the special case were an extra _#_ is used (e.g. _1_ in file_1_5.93)
        ## we remove it.
        count = 0
        for item in filename_list:
            if item.isdigit():
                ## e.g., file_1_5.39-3.tar.gz -> ['file', '1', '5.39-3.tar.gz']
                del filename_list[count]   # delete the item (e.g, '1')
                ## file_1_5.39-3.tar.gz -> ['file','5.39-3.tar.gz']
                break  # quit after finding the first one (e.g. _1_ in file_1_5.93)
            count += 1
    else:  ## I have no idea
        return ('', '')

    num_subparts = len (filename_list)
    display_name = ''
    count = 0
    for subpart in filename_list:
    ## for each subpart
        count += 1
        if count == 1: ## for FIRST subpart
            ## next check when only one sub part - e.g., no version present
            if num_subparts == 1: ## e.g.,  myprogram.zip, utils.tar.gz app.1.2.zip
                print ("only one subpart:", subpart, filename_list, file=sys.stderr )
                ## 'app.zip' --> ['.zip']  utils.tar.gz --> ['.tar', '.gz']   velero.1.2.zip --> ['.1', '.2', '.zip']
                extensions_list = Path(subpart).suffixes
                if len(extensions_list) == 0: ##  myapp (rare if at all)
                    display_name += subpart + ':'
                elif len(extensions_list) == 1:  ## 'app.zip' --> ['.zip']
                    index = subpart.find (extensions_list[0])
                    display_name += subpart[:index] + ':'
                elif len(extensions_list) >= 2 and '.tar' in extensions_list: ## utils.tar.gz --> ['.tar', '.gz'], app.1.2.tar.gz --> ['.1', '.2', '.tar', '.gz']
                    ##index = subpart.find (extensions_list[0] + extensions_list[1])
                    all_extensions = ''.join(extensions_list) ## --> '.1.2.tar.gz'
                    ext1 = extensions_list.pop()
                    ext2 = extensions_list.pop()
                    remaing_extensions = ''.join(extensions_list) ## --> '.1.2'
                    print ("extx", ext1, ext2, extensions_list, display_name, file=sys.stderr)
                    index = subpart.find (all_extensions) ## index for '.1.2.tar.gz'
                    display_name +=  subpart[:index] + ':' + remaing_extensions [1:] ## [1:] removes '.' from '.1.2' --> --> '1.2'
                    print ("here:", display_name, "|"+ ext2 + ext1 + "|", subpart[:index], file=sys.stderr)
                else: ## > 2 two where second is not '.tar'   e.g.,  app.1.2.3.gz --> ['.1', '.2', '.3', '.gz']  --> ('app','1.2.3')
                    extensions_list.pop() ## pop '.gz'
                    remaing_extensions = ''.join(extensions_list) ## --> '.1.2.3'
                    index = subpart.find (remaing_extensions) ## index for '.1.2.3'
                    display_name += subpart[:index] + ':' + remaing_extensions [1:] ## [1:] removes '.' from '.1.2.3' --> '1.2.3'
                found = True
                break
            else:
                ## two or more subparts - assume first subpart is part of component name
                display_name += subpart
            continue

        ## Work on the version part of the file name
        if '.' in subpart or subpart[0].isdigit():  ## if '.' or first char is digit
            found_verion = False
            ## remove odd strings if any found.
            for odd_str in odd_str_list:
                index = subpart.find (odd_str)
                if index != -1:
                    ## slice off odd_str
                    subpart_slice =  subpart[:index]
                    subpart = subpart_slice

            ## next check if this is the last subpart
            if count == num_subparts and not has_digits(subpart):  ## if last subpart and no digits
                ## no version found in last subpart (i.e., no digits)
                display_name += "-" + subpart + ':'
            else: ## not last but found version OR last piece with version
                display_name += ':' + subpart
            found_verion = True

            if found_verion == False:
                display_name += ':' + subpart
            break
        else: ## '.' not found and no digits
            display_name += "-" + subpart ## continue adding subparts
    ## display name consists of component name + version seperated by a ':'. Obtain both
    name_version_list = display_name.split (':')

    if len (name_version_list) > 1:
        name = name_version_list[0]
        version = name_version_list[1]
    elif len (name_version_list) == 1:
        name = name_version_list[0]
        version = ""
    else:
        name = ''
        version = ''

    ## When '-' is the main delimeter but there is a '_' BEFORE version we need to clean it up
    ## For example:
    ##                                           name            version
    ##  init-system-helpers_1.60.tar.gz   --> ['init-system', 'helpers_1.60']
    ##    we want                         --> ['init-system-helpers', '1.60']
    if dash_delimiter_used and '_' in version:
        ## split up version using the '_' delimeter and adding first part back to the name
        version_parts = version.split ('_')
        name += '-' + version_parts[0]
        version = version_parts[1]


    ## Some files have "+gitautoinc+" in their name. We need to remove it.
    ## for example: azure-uhttp-c-lts_07_2020+gitAUTOINC+6f18cb8e7f-r0-p019b374.tar.gz
    ## See if exists if index >= 0
    # str_index = name.find("+gitautoinc+")
    # if str_index > -1:
    #     ## found substring. Remove it.
    #     name = name[:str_index]

    return (name, version)

    # if len (name_version_list) > 1:
    #     return (name_version_list[0], name_version_list[1])
    # elif len (name_version_list) == 1:
    #     return (name_version_list[0], '')
    # else:
    #     return ('', '')




def license_translate (the_string):
    pass